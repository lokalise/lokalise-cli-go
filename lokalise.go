package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lokalise/lokalise-cli-go/lokalise"
	"github.com/urfave/cli"
)

func main() {
	var apiToken string
	var configFile string

	type Config struct {
		Token   string
		Project string
	}

	app := cli.NewApp()
	app.Name = "Lokalise CLI tool"
	app.Version = "v0.49"
	app.Compiled = time.Now()
	app.Usage = "upload and download language files."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "token",
			Usage:       "API `token` is required and can be obtained under your Account page in Lokalise.",
			Destination: &apiToken,
		},
		cli.StringFlag{
			Name:        "config",
			Usage:       "Load configuration from `file`. Looks up /etc/lokalise.cfg by default.",
			Destination: &configFile,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List your projects at Lokalise.",
			Action: func(c *cli.Context) error {
				var conf Config

				if configFile == "" {
					configFile = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(configFile, &conf); err != nil {
					// do nothing if no config
				}

				if apiToken == "" {
					apiToken = conf.Token
				}

				if apiToken == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				projects, err := lokalise.List(apiToken)
				if err != nil {
					fmt.Printf("%v\n", err)
					return cli.NewExitError("ERROR: API returned error (see above)", 7)
				}

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)
				cRed := color.New(color.FgRed)
				cCyan := color.New(color.FgCyan)

				for _, project := range projects {
					cWhite.Print(project.ID)
					if project.Owner == "1" {
						cGreen.Print(" (admin) ")
					} else {
						cRed.Print(" (contr) ")
					}
					cCyan.Print(" ", project.Name)
				}

				return nil
			},
		},
		{
			Name:    "export",
			Aliases: []string{"d"},
			Usage:   "Downloads language files.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type",
					Usage: "File format to export. See https://lokalise.co/apidocs#file_formats (single value, required)",
				},
				cli.StringFlag{
					Name:  "dest",
					Usage: "Destination directory on local filesystem (for the .zip bundle). (/dir)",
				},
				cli.StringFlag{
					Name:  "unzip_to",
					Usage: "Unzip downloaded bundle to a specified directory and remove the .zip. Use --keep_zip to avoid the deletion. (/dir)",
				},
				cli.StringFlag{
					Name:  "keep_zip",
					Usage: "Keep downloaded .zip, if --unzip_to is used. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "langs",
					Usage: "Languages to include. Don't specify for all languages. (comma separated)",
				},
				cli.StringFlag{
					Name:  "use_original",
					Usage: "Use original filenames/formats. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "filter",
					Usage: "Filter by 'translated', 'nonfuzzy', 'nonhidden' fields.(comma separated)",
				},
				cli.StringFlag{
					Name:  "bundle_structure",
					Usage: ".ZIP bundle structure (see docs.lokalise.co for placeholders).",
				},
				cli.StringFlag{
					Name:  "webhook_url",
					Usage: "Sends POST['file'] if specified. (url)",
				},
				cli.StringFlag{
					Name:  "export_all",
					Usage: "Include all platform keys. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "export_empty",
					Usage: "How to export empty strings. (empty, base, skip)",
				},
				cli.StringFlag{
					Name:  "include_comments",
					Usage: "Include comments in exported file. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "include_pids",
					Usage: "Other projects ID's, which keys to include in this export. (comma separated)",
				},
				cli.StringFlag{
					Name:  "tags",
					Usage: "Filter keys by tags (comma separated)",
				},
				cli.StringFlag{
					Name:  "yaml_include_root",
					Usage: "Include language ISO code as root key in YAML export. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "json_unescaped_slashes",
					Usage: "Leave forward slashes unescaped in JSON export. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "export_sort",
					Usage: "Key sort order (first_added, last_added, last_updated, a_z, z_a)",
				},
				cli.StringFlag{
					Name:  "replace_breaks",
					Usage: "Replace link breaks with \\n. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "no_language_folders",
					Usage: "Don't use language folders. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "triggers",
					Usage: "Trigger integration export. Allowed values are 'amazons3' and 'gcs'. (comma separated)",
				},
			},
			Action: func(c *cli.Context) error {
				var conf Config

				if configFile == "" {
					configFile = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(configFile, &conf); err != nil {
					// do nothing if no config
				}

				if apiToken == "" {
					apiToken = conf.Token
				}
				if apiToken == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				projectID := c.Args().First()
				if projectID == "" {
					projectID = conf.Project
				}
				if projectID == "" {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help export` for all options.", 5)
				}

				fileType := c.String("type")
				if fileType == "" {
					return cli.NewExitError("ERROR: --type is required. Run `lokalise help export` for all options.", 5)
				}

				dest := c.String("dest")
				if dest == "" {
					dest = "."
				}
				opts := lokalise.ExportOptions{
					UseOriginal:          optionalBool(c.String("use_original")),
					BundleStructure:      optionalString(c.String("bundle_structure")),
					WebhookURL:           optionalString(c.String("webhook_url")),
					ExportAll:            optionalBool(c.String("export_all")),
					ExportEmpty:          optionalString(c.String("export_empty")),
					ExportSort:           optionalString(c.String("export_sort")),
					IncludeComments:      optionalBool(c.String("include_comments")),
					ReplaceBreaks:        optionalBool(c.String("replace_breaks")),
					YAMLIncludeRoot:      optionalBool(c.String("yaml_include_root")),
					JSONUnescapedSlashes: optionalBool(c.String("json_unescaped_slashes")),
					NoLanguageFolders:    optionalBool(c.String("no_language_folders")),
					Languages:            commaSlice(c.String("langs")),
					Filter:               commaSlice(c.String("filter")),
					Triggers:             commaSlice(c.String("triggers")),
					IncludePIDs:          commaSlice(c.String("include_pids")),
					Tags:                 commaSlice(c.String("tags")),
				}

				unzipTo := c.String("unzip_to")
				keepZip := c.String("keep_zip")
				if keepZip == "" {
					keepZip = "0"
				}

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)

				theSpinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
				theSpinner.Start()
				fmt.Print("Requesting...")

				bundle, err := lokalise.Export(apiToken, projectID, fileType, &opts)
				theSpinner.Stop()
				if err != nil {
					fmt.Printf("\n%v\n", err)
					return cli.NewExitError("ERROR: API returned error (see above)", 7)
				}

				filename := strings.Split(bundle.File, "/")[4]

				cWhite.Println()
				cWhite.Print("Remote ")
				cGreen.Print(bundle.FullFile + "... ")
				cWhite.Println("OK")

				cWhite.Print("Local ")
				cGreen.Print(path.Join(dest, filename) + "... ")

				downloadFile(path.Join(dest, filename), bundle.FullFile)
				cWhite.Println("OK")

				if unzipTo != "" {
					files, err := unzip(path.Join(dest, filename), unzipTo)

					if err != nil {
						cWhite.Println("Error unzipping files")
					} else {
						cWhite.Print("Unzipped ")
						cGreen.Print(strings.Join(files, ", ") + " ")
						cWhite.Println("OK")
						if keepZip == "0" {
							os.Remove(path.Join(dest, filename))
						}
					}

				}

				return nil
			},
		},
		{
			Name:    "import",
			Usage:   "Upload language files.",
			Aliases: []string{"u"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "A single file, or a comma-separated list of files or file masks on the local filesystem to import (any of the supported file formats) (required). Make sure to escape * if using file masks (\\*).",
				},
				cli.StringFlag{
					Name:  "lang_iso",
					Usage: "Language of the translations in the file being imported. Applies to all files, if using a list of a file mask. (reqired)",
				},
				cli.StringFlag{
					Name:  "replace",
					Usage: "Shall existing translations be replaced. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "fill_empty",
					Usage: "If values are empty, keys will be copied to values. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "distinguish",
					Usage: "Distinguish similar keys in different files. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "hidden",
					Usage: "Hide imported keys from contributors. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "tags",
					Usage: "Tags list for newly imported keys. (comma separated)",
				},
				cli.StringFlag{
					Name:  "use_trans_mem",
					Usage: "Use translation memory to fill 100% matches. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "replace_breaks",
					Usage: "Replace \\n with line breaks. (`0/1`)",
				},
			},
			Action: func(c *cli.Context) error {
				var conf Config

				if configFile == "" {
					configFile = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(configFile, &conf); err != nil {
					// do nothing if no config
				}

				if apiToken == "" {
					apiToken = conf.Token
				}
				if apiToken == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				projectID := c.Args().First()
				if projectID == "" {
					projectID = conf.Project
				}
				if projectID == "" {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help import` for all options.", 5)
				}

				file := c.String("file")
				if file == "" {
					return cli.NewExitError("ERROR: --file required.  Run `lokalise help import` for all options.", 5)
				}

				langIso := c.String("lang_iso")
				if langIso == "" {
					return cli.NewExitError("ERROR: --lang_iso is required. If you are using filemask in --file parameter, make sure escape it (e.g. \\*.json).  Run `lokalise help import` for all options. ", 5)
				}

				importOptions := lokalise.ImportOptions{
					Replace:       optionalBool(c.String("replace")),
					FillEmpty:     optionalBool(c.String("fill_empty")),
					Distinguish:   optionalBool(c.String("distinguish")),
					Hidden:        optionalBool(c.String("hidden")),
					Tags:          commaSlice(c.String("tags")),
					ReplaceBreaks: optionalBool(c.String("replace_breaks")),
				}

				// FIXME: undocumented
				// useTransMem := c.String("use_trans_mem")

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)

				fileMasks := strings.Split(file, ",")
				for _, mask := range fileMasks {
					files, err := filepath.Glob(mask)
					if err != nil {
						return cli.NewExitError("ERROR: file glob pattern not valid", 5)
					}

					for _, filename := range files {
						theSpinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
						cWhite.Printf("Uploading %s... ", filename)
						theSpinner.Start()
						result, err := lokalise.Import(apiToken, projectID, file, langIso, &importOptions)
						theSpinner.Stop()
						if err != nil {
							fmt.Printf("\n%v\n", err)
							return cli.NewExitError("ERROR: API returned error (see above)", 7)
						}
						cGreen.Print("Inserted ")
						cWhite.Print(result.Inserted)
						cGreen.Print(", skipped ")
						cWhite.Print(result.Skipped)
						cGreen.Print(", updated ")
						cWhite.Print(result.Updated)
						cGreen.Println(" keys.")
					}
				}

				return nil
			},
		},
	}

	app.Run(os.Args)
}

func downloadFile(filepath string, url string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func commaSlice(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

func optionalBool(v string) *bool {
	if v == "" {
		return nil
	}
	var b bool
	if v == "1" {
		b = true
	}
	return &b
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func unzip(src, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				return filenames, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return filenames, err
			}
		}
	}
	return filenames, nil
}
