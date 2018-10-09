package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
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
	app.Version = "v0.66"
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
					cCyan.Println(" ", project.Name)
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
					Usage: "Filter by 'translated', 'reviewed', 'nonfuzzy', 'nonhidden' fields. (comma separated)",
				},
				cli.StringFlag{
					Name:  "bundle_structure",
					Usage: "Bundle file structure (use with --use_original=0). See https://lokalise.co/apidocs#export",
				},
				cli.StringFlag{
					Name:  "directory_prefix",
					Usage: "Directory prefix in the bundle (use with --use_original=1). See https://lokalise.co/apidocs#export",
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
					Usage: "Depreacted. Use include_tags instead: Only include keys with these tags (comma separated)",
				},
				cli.StringFlag{
					Name:  "include_tags",
					Usage: "Only include keys with these tags (comma separated)",
				},
				cli.StringFlag{
					Name:  "exclude_tags",
					Usage: "Do not include keys with these tags (comma separated)",
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
					Name:  "java_properties_encoding",
					Usage: "Encoding for .properties files. (utf-8, latin-1)",
				},
				cli.StringFlag{
					Name:  "java_properties_separator",
					Usage: "Separator for keys/values in .properties files. (=, :)",
				},
				cli.StringFlag{
					Name:  "export_sort",
					Usage: "Key sort order. (first_added, last_added, last_updated, a_z, z_a)",
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
					Usage: "Trigger integration export. Allowed values are 'amazons3', 'gcs', 'gitlab', 'github', 'bitbucket'. (comma separated)",
				},
				cli.StringFlag{
					Name:  "repos",
					Usage: "If a repo integration is triggered, specify to which repos the pull requests should go to. Don't specify for all. (comma separated)",
				},
				cli.StringFlag{
					Name:  "plural_format",
					Usage: "Override default plural format. See https://lokalise.co/apidocs#pl_ph_formats (value).",
				},
				cli.StringFlag{
					Name:  "icu_numeric",
					Usage: "Use =0, =1, =2 instead of zero, one, two plural forms. Works with ICU plurals only. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "placeholder_format",
					Usage: "Override default placeholder format. See https://lokalise.co/apidocs#pl_ph_formats (value).",
				},
				cli.StringFlag{
					Name:  "indentation",
					Usage: "Provide to override default indentation in supported files. (1sp, 2sp, 3sp, 4sp, 5sp, 6sp, 7sp, 8sp, tab)",
				},
				cli.StringFlag{
					Name:  "escape_percent",
					Usage: "When enabled, all universal percent placeholders [%] will be always exported as %%. Only works for printf placeholder format ('0/1').",
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

				// map legacy flags to new names
				if legacyTags := c.String("tags"); len(legacyTags) != 0 {
					color.New(color.FgRed).Println("WARNING: --tags is deprecated. Use --include_tags instead.")
					c.Set("include_tags", legacyTags)
				}

				var opts []lokalise.ExportOption
				opts = setExportBool(opts, c, "use_original", lokalise.WithOriginal)
				opts = setExportString(opts, c, "bundle_structure", lokalise.WithBundleStructure)
				opts = setExportString(opts, c, "directory_prefix", lokalise.WithDirectoryPrefix)
				opts = setExportString(opts, c, "webhook_url", lokalise.WithWebhookURL)
				opts = setExportBool(opts, c, "export_all", lokalise.WithAll)
				opts = setExportString(opts, c, "export_empty", lokalise.WithEmpty)
				opts = setExportString(opts, c, "export_sort", lokalise.WithSortOrder)
				opts = setExportString(opts, c, "java_properties_encoding", lokalise.WithJavaPropertiesEncoding)
				opts = setExportString(opts, c, "java_properties_separator", lokalise.WithJavaPropertiesSeparator)
				opts = setExportString(opts, c, "placeholder_format", lokalise.WithPlaceholderFormat)
				opts = setExportString(opts, c, "indentation", lokalise.WithIndentation)
				opts = setExportString(opts, c, "plural_format", lokalise.WithPluralFormat)
				opts = setExportBool(opts, c, "include_comments", lokalise.WithComments)
				opts = setExportBool(opts, c, "replace_breaks", lokalise.WithExportReplaceBreaks)
				opts = setExportBool(opts, c, "yaml_include_root", lokalise.WithYAMLRoot)
				opts = setExportBool(opts, c, "json_unescaped_slashes", lokalise.WithJSONUnescapedSlashes)
				opts = setExportBool(opts, c, "no_language_folders", lokalise.WithNoLanguageFolders)
				opts = setExportBool(opts, c, "icu_numeric", lokalise.WithICUNumeric)
				opts = setExportBool(opts, c, "escape_percent", lokalise.WithPercentEscape)
				opts = setExportStrings(opts, c, "langs", lokalise.WithLanguages)
				opts = setExportStrings(opts, c, "filter", lokalise.WithFilter)
				opts = setExportStrings(opts, c, "triggers", lokalise.WithTriggers)
				opts = setExportStrings(opts, c, "repos", lokalise.WithRepos)
				opts = setExportStrings(opts, c, "include_pids", lokalise.WithPIDs)
				opts = setExportStrings(opts, c, "include_tags", lokalise.WithIncludeTags)
				opts = setExportStrings(opts, c, "exclude_tags", lokalise.WithExcludeTags)

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

				bundle, err := lokalise.Export(apiToken, projectID, fileType, opts...)
				theSpinner.Stop()
				if err != nil {
					fmt.Printf("\n%v\n", err)
					return cli.NewExitError("ERROR: API returned error (see above)", 7)
				}

				if bundle.File != "" {
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
				} else {
					cWhite.Println("OK")
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
					Name:  "convert_placeholders",
					Usage: "Convert placeholders to Lokalise universal ones. https://docs.lokalise.co/developer-docs/universal-placeholders (`0/1`)",
				},
				cli.StringFlag{
					Name:  "fill_empty",
					Usage: "If values are empty, keys will be copied to values. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "icu_plurals",
					Usage: "Enable to automatically detect and parse ICU formatted plurals. (`0/1`)",
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
					Name:  "include_path",
					Usage: "Include relative directory name in the filename when uploading. Do not enable if path contains language code. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "replace_breaks",
					Usage: "Replace \\n with line breaks. (`0/1`)",
				},
				cli.StringFlag{
					Name:  "cleanup_mode",
					Usage: "Enable to delete keys with all language translations from Lokalise that are not present in the uploaded files. (`0/1`)",
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

				includePath, _ := strconv.ParseBool(c.String("include_path"))

				var opts []lokalise.ImportOption
				opts = setImportBool(opts, c, "replace", lokalise.WithReplace)
				opts = setImportBool(opts, c, "convert_placeholders", lokalise.WithConvertPlaceholders)
				opts = setImportBool(opts, c, "icu_plurals", lokalise.WithICUPlurals)
				opts = setImportBool(opts, c, "fill_empty", lokalise.WithFillEmpty)
				opts = setImportBool(opts, c, "distinguish", lokalise.WithDistinguish)
				opts = setImportBool(opts, c, "hidden", lokalise.WithHidden)
				opts = setImportBool(opts, c, "use_trans_mem", lokalise.WithTranslationMemory)
				opts = setImportStrings(opts, c, "tags", lokalise.WithTags)
				opts = setImportBool(opts, c, "replace_breaks", lokalise.WithImportReplaceBreaks)
				opts = setImportBool(opts, c, "cleanup_mode", lokalise.WithCleanupMode)

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)

				fileMasks := strings.Split(file, ",")
				for _, mask := range fileMasks {
					files, err := filepath.Glob(mask)
					if err != nil {
						return cli.NewExitError("ERROR: file glob pattern not valid", 5)
					}

					for _, filename := range files {
						if includePath {
							opts = append(opts, lokalise.WithFilename(filename))
						}
						theSpinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
						cWhite.Printf("Uploading %s... ", filename)
						theSpinner.Start()
						result, err := lokalise.Import(apiToken, projectID, filename, langIso, opts...)
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

func setExportBool(opts []lokalise.ExportOption, c *cli.Context, cmdField string, f func(v bool) lokalise.ExportOption) []lokalise.ExportOption {
	value := c.String(cmdField)
	if value == "" {
		return opts
	}
	b, _ := strconv.ParseBool(value)
	return append(opts, f(b))
}
func setExportString(opts []lokalise.ExportOption, c *cli.Context, cmdField string, f func(v string) lokalise.ExportOption) []lokalise.ExportOption {
	value := c.String(cmdField)
	if value == "" {
		return opts
	}
	return append(opts, f(value))
}
func setExportStrings(opts []lokalise.ExportOption, c *cli.Context, cmdField string, f func(v ...string) lokalise.ExportOption) []lokalise.ExportOption {
	value := commaSlice(c.String(cmdField))
	if len(value) == 0 {
		return opts
	}
	return append(opts, f(value...))
}

func commaSlice(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

func setImportBool(opts []lokalise.ImportOption, c *cli.Context, cmdField string, f func(v bool) lokalise.ImportOption) []lokalise.ImportOption {
	value := c.String(cmdField)
	if value == "" {
		return opts
	}
	b, _ := strconv.ParseBool(value)
	return append(opts, f(b))
}
func setImportString(opts []lokalise.ImportOption, c *cli.Context, cmdField string, f func(v string) lokalise.ImportOption) []lokalise.ImportOption {
	value := c.String(cmdField)
	if len(value) == 0 {
		return opts
	}
	return append(opts, f(value))
}
func setImportStrings(opts []lokalise.ImportOption, c *cli.Context, cmdField string, f func(v ...string) lokalise.ImportOption) []lokalise.ImportOption {
	value := commaSlice(c.String(cmdField))
	if len(value) == 0 {
		return opts
	}
	return append(opts, f(value...))
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
