package main

import (
	"os"
	"bytes"
	"net/http"
	"github.com/urfave/cli"
	"log"
	"io/ioutil"
	"time"
	"encoding/json"
	"github.com/fatih/color"
	"fmt"
	"strings"
	"io"
	"github.com/briandowns/spinner"
	"mime/multipart"
	"path/filepath"
	"archive/zip"
	"github.com/BurntSushi/toml"
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
	app.Version = "v0.47"
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
			Name:  "list",
			Aliases: []string{"l"},
			Usage: "List your projects at Lokalise.",
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

				res, err := http.Get("https://api.lokalise.co/api/project/list?api_token=" + apiToken)
				if err != nil {
					log.Fatal(err)
				}

				response, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					log.Fatal(err)
				}

				var dat map[string]interface{}
				if err := json.Unmarshal([]byte(response), &dat); err != nil {
					panic(err)
				}

				status := dat["response"].(map[string]interface{})["status"]

				if status == "error" {
					e := dat["response"].(map[string]interface{})["message"]
					fmt.Println(e)
					return cli.NewExitError("ERROR: API returned error (see above)", 7)
				}

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)
				cRed := color.New(color.FgRed)
				cCyan := color.New(color.FgCyan)

				for _, v := range dat {
					switch vv := v.(type) {
					case []interface{}:
						for _, u := range vv {
							project := u.(map[string]interface{})

							cWhite.Print(project["id"])
							if project["owner"] == "1" {
								cGreen.Print(" (admin) ")
							} else {
								cRed.Print(" (contr) ")
							}
							cCyan.Println(" ", project["name"])
						}
					}
				}

				return nil
			},
		},
		{
			Name:  "export",
			Aliases: []string{"d"},
			Usage: "Downloads language files.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "type",
					Usage: "File format (strings, xliff, plist, xml, properties, json, po, php, ini, yml, xls) (single value, required)",
				},
				cli.StringFlag{
					Name: "dest",
					Usage: "Destination directory on local filesystem (for the .zip bundle) (/dir)",
				},
				cli.StringFlag{
					Name: "unzip_to",
					Usage: "Unzip downloaded bundle to specified directory (/dir)",
				},
				cli.StringFlag{
					Name: "langs",
					Usage: "Languages to include. Don't specify for all languages. (comma separated)",
				},
				cli.StringFlag{
					Name: "use_original",
					Usage: "Use original filenames/formats (`0/1`)",
				},
				cli.StringFlag{
					Name: "filter",
					Usage: "Filter by 'translated', 'nonfuzzy', 'nonhidden' fields (comma separated)",
				},
				cli.StringFlag{
					Name: "bundle_structure",
					Usage: ".ZIP bundle structure (see docs.lokalise.co for placeholders)",
				},
				cli.StringFlag{
					Name: "webhook_url",
					Usage: "Sends POST['file'] if specified (url)",
				},
				cli.StringFlag{
					Name: "export_all",
					Usage: "Include all platform keys (`0/1`)",
				},
				cli.StringFlag{
					Name: "ota_plugin_bundle",
					Usage: "Generate plugin for OTA iOS plugin (`0/1`)",
				},
				cli.StringFlag{
					Name: "export_empty",
					Usage: "How to export empty strings (empty, base, skip)",
				},
				cli.StringFlag{
					Name: "include_comments",
					Usage: "Include comments in exported file (`0/1`)",
				},
				cli.StringFlag{
					Name: "include_pids",
					Usage: "Other projects ID's, which keys to include in this export (comma separated)",
				},
				cli.StringFlag{
					Name: "tags",
					Usage: "Filter keys by tags (comma separated)",
				},
				cli.StringFlag{
					Name: "yaml_include_root",
					Usage: "Include language ISO code as root key in YAML export (`0/1`)",
				},
				cli.StringFlag{
					Name: "json_unescaped_slashes",
					Usage: "Leave forward slashes unescaped in JSON export (`0/1`)",
				},
				cli.StringFlag{
					Name: "export_sort",
					Usage: "Key sort order (first_added, last_added, last_updated, a_z, z_a)",
				},
				cli.StringFlag{
					Name: "replace_breaks",
					Usage: "Replace link breaks with \\n (`0/1`)",
				},
				cli.StringFlag{
					Name: "no_language_folders",
					Usage: "Don't use language folders (`0/1`)",
				},
				cli.StringFlag{
					Name: "triggers",
					Usage: "Trigger integration export. Allowed values are 'amazons3' and 'gcs' (comma separated)",
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

				projectId := c.Args().First();
				if (projectId == "") {
					projectId = conf.Project
				}
				if (projectId == "") {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help export` for all options.", 5)
				}

				dest := c.String("dest")
				if dest == "" {
					dest = "."
				}

				fileType := c.String("type")
				if fileType == "" {
					return cli.NewExitError("ERROR: --type is required. Run `lokalise help export` for all options.", 5)
				}

				theUrl := "api_token=" + apiToken + `&id=` + projectId + `&type=` + fileType

				langs := c.String("langs")
				if (langs != "") {
					langs = s2json(langs)
					theUrl += "&langs=" + langs
				}

				useOriginal := c.String("use_original")
				if (useOriginal != "") {
					theUrl += "&use_original=" + useOriginal
				}

				noLanguageFolders := c.String("no_language_folders")
				if (noLanguageFolders != "") {
					theUrl += "&switch_no_language_folders=" + noLanguageFolders
				}

				filter := c.String("filter")
				if (filter != "") {
					filter = s2json(filter)
					theUrl += "&filter=" + filter
				}

				triggers := c.String("triggers")
				if (triggers != "") {
					triggers = s2json(triggers)
					theUrl += "&triggers=" + triggers
				}

				bundleStructure := c.String("bundle_structure")
				if (bundleStructure != "") {
					theUrl += "&bundle_structure=" + bundleStructure
				}

				webhookUrl := c.String("webhook_url")
				if (webhookUrl != "") {
					theUrl += "&webhook_url=" + webhookUrl
				}

				exportAll := c.String("export_all")
				if (exportAll != "") {
					theUrl += "&export_all=" + exportAll
				}

				otaPluginBundle := c.String("ota_plugin_bundle")
				if (otaPluginBundle != "") {
					theUrl += "&ota_plugin_bundle=" + otaPluginBundle
				}

				exportEmpty := c.String("export_empty")
				if (exportEmpty != "") {
					theUrl += "&export_empty=" + exportEmpty
				}

				exportSort := c.String("export_sort")
				if (exportSort != "") {
					theUrl += "&export_sort=" + exportSort
				}

				includeComments := c.String("include_comments")
				if (includeComments != "") {
					theUrl += "&include_comments=" + includeComments
				}

				replaceBreaks := c.String("replace_breaks")
				if (replaceBreaks != "") {
					theUrl += "&replace_breaks=" + replaceBreaks
				}

				yamlIncludeRoot := c.String("yaml_include_root")
				if (yamlIncludeRoot != "") {
					theUrl += "&yaml_include_root=" + yamlIncludeRoot
				}

				jsonUnescapedSlashes := c.String("json_unescaped_slashes")
				if (jsonUnescapedSlashes != "") {
					theUrl += "&json_unescaped_slashes=" + jsonUnescapedSlashes
				}

				includePids := c.String("include_pids")
				if (includePids != "") {
					includePids = s2json(includePids)
					theUrl += "&include_pids=" + includePids
				}

				tags := c.String("tags")
				if (tags != "") {
					tags = s2json(tags)
					theUrl += "&tags=" + tags
				}

				unzipTo := c.String("unzip_to")

				theSpinner := spinner.New(spinner.CharSets[9], 100 * time.Millisecond)
				theSpinner.Start()
				fmt.Print("Requesting...")
				body := strings.NewReader(theUrl)
				req, err := http.NewRequest("POST", "https://api.lokalise.co/api/project/export", body)
				if err != nil {
					log.Fatal(err)
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := http.DefaultClient.Do(req)

				if err != nil {
					log.Fatal(err)
				}

				response, err := ioutil.ReadAll(resp.Body)
				defer resp.Body.Close()

				var dat map[string]interface{}
				if err := json.Unmarshal([]byte(response), &dat); err != nil {
					panic(err)
				}

				var file string
				file = ""

				status := dat["response"].(map[string]interface{})["status"]

				if status == "error" {
					e := dat["response"].(map[string]interface{})["message"]
					fmt.Println(e)
					return cli.NewExitError("ERROR: API returned error (see above)", 7)
				} else {
					file = dat["bundle"].(map[string]interface{})["file"].(string)
				}

				fileUrl := "https://s3-eu-west-1.amazonaws.com/lokalise-assets/" + file
				filename := strings.Split(file, "/")[4]

				theSpinner.Stop()

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)

				cWhite.Println()
				cWhite.Print("Remote ")
				cGreen.Print(fileUrl + "... ")
				cWhite.Println("OK")

				cWhite.Print("Local ")
				cGreen.Print(dest + "/" + filename + "... ")

				downloadFile(dest + "/" + filename, fileUrl)
				cWhite.Println("OK")

				if (unzipTo != "") {
					files, err := Unzip(dest + "/" + filename, unzipTo)

					if err != nil {
						cWhite.Println("Error unzipping files")
					} else {
						cWhite.Print("Unzipped ")
						cGreen.Print(strings.Join(files, ", ") + " ")
						cWhite.Println("OK")
					}

				}

				return nil
			},
		},
		{
			Name:  "import",
			Usage: "Upload language files.",
			Aliases: []string{"u"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "file",
					Usage: "A single file, or comma-separated list of files or file masks on local filesystem to import (any of the supported file formats) (required). Make sure to escape * if using file masks (\\*).",
				},
				cli.StringFlag{
					Name: "lang_iso",
					Usage: "Language of the translations being imported (reqired)",
				},
				cli.StringFlag{
					Name: "replace",
					Usage: "Shall existing translations be replaced (`0/1`)",
				},
				cli.StringFlag{
					Name: "fill_empty",
					Usage: "If values are empty, keys will be copied to values (`0/1`)",
				},
				cli.StringFlag{
					Name: "distinguish",
					Usage: "Distinguish similar keys in different files (`0/1`)",
				},
				cli.StringFlag{
					Name: "hidden",
					Usage: "Hide imported keys from contributors (`0/1`)",
				},
				cli.StringFlag{
					Name: "tags",
					Usage: "Tags list for newly imported keys (comma separated)",
				},
				cli.StringFlag{
					Name: "use_trans_mem",
					Usage: "Use translation memory to fill 100% matches (`0/1`)",
				},
				cli.StringFlag{
					Name: "replace_breaks",
					Usage: "Replace \\n with line breaks (`0/1`)",
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

				theUrl := "https://api.lokalise.co/api/project/import"

				projectId := c.Args().First();
				if (projectId == "") {
					projectId = conf.Project
				}
				if (projectId == "") {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help import` for all options.", 5)
				}

				file := c.String("file")
				if (file == "") {
					return cli.NewExitError("ERROR: --file required.  Run `lokalise help import` for all options.", 5)
				}

				langIso := c.String("lang_iso")
				if langIso == "" {
					return cli.NewExitError("ERROR: --lang_iso is required. If you are using filemask in --file parameter, make sure escape it (e.g. \\*.json).  Run `lokalise help import` for all options. ", 5)
				}

				fillEmpty := c.String("fill_empty")
				hidden := c.String("hidden")
				distinguish := c.String("distinguish")
				replace := c.String("replace")
				useTransMem := c.String("use_trans_mem")
				replaceBreaks := c.String("replace_breaks")

				tags := c.String("tags")
				if (tags != "") {
					tags = s2json(tags)
				}

				extraParams := map[string]string{
					"id": projectId,
					"api_token": apiToken,
					"lang_iso": langIso,
					"fill_empty": fillEmpty,
					"hidden": hidden,
					"distinguish": distinguish,
					"replace": replace,
					"tags": tags,
					"use_trans_mem": useTransMem,
					"replace_breaks": replaceBreaks,
				}

				fileMasks := strings.Split(file, ",")

				cWhite := color.New(color.FgHiWhite)
				cGreen := color.New(color.FgGreen)

				for _, mask := range fileMasks {
					files, err := filepath.Glob(mask)

					if err != nil {
						log.Fatal(err)
					}

					for _, filename := range files {
						theSpinner := spinner.New(spinner.CharSets[9], 100 * time.Millisecond)
						theSpinner.Start()

						cWhite.Print("Uploading ")
						cWhite.Print(filename)
						cWhite.Print("... ")

						request, err := newfileUploadRequest(theUrl, extraParams, "file", filename)

						if err != nil {
							log.Fatal(err)
						}
						client := &http.Client{}
						resp, err := client.Do(request)
						if err != nil {
							log.Fatal(err)
						}
						theSpinner.Stop()
						response, err := ioutil.ReadAll(resp.Body)
						defer resp.Body.Close()

						var dat map[string]interface{}
						if err := json.Unmarshal([]byte(response), &dat); err != nil {
							panic(err)
						}

						status := dat["response"].(map[string]interface{})["status"]

						if status == "error" {
							e := dat["response"].(map[string]interface{})["message"]
							fmt.Println(e)
							return cli.NewExitError("ERROR: API returned error (see above)", 7)
						} else {
							skipped, _ := dat["result"].(map[string]interface{})["skipped"].(float64)
							inserted, _ := dat["result"].(map[string]interface{})["inserted"].(float64)
							updated, _ := dat["result"].(map[string]interface{})["updated"].(float64)

							cGreen.Print("Inserted ")
							cWhite.Print(inserted)
							cGreen.Print(", skipped ")
							cWhite.Print(skipped)
							cGreen.Print(", updated ")
							cWhite.Print(updated)
							cGreen.Println(" keys.")
						}
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

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", uri, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	return req, nil
}

func s2json(s string) string {
	return "['" + strings.Join(strings.Split(s, ","), "','") + "']"
}

func Unzip(src, dest string) ([]string, error) {
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
				log.Fatal(err)
				return filenames, err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, f.Mode())
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