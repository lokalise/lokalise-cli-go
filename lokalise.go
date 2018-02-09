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
	var api_token string
	var config_file string

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
			Destination: &api_token,
		},
		cli.StringFlag{
			Name:        "config",
			Usage:       "Load configuration from `file`. Looks up /etc/lokalise.cfg by default.",
			Destination: &config_file,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "list",
			Aliases: []string{"l"},
			Usage: "List your projects at Lokalise.",
			Action: func(c *cli.Context) error {
				var conf Config

				if config_file == "" {
					config_file = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(config_file, &conf); err != nil {
					// do nothing if no config
				}

				if api_token == "" {
					api_token = conf.Token
				}

				if api_token == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				res, err := http.Get("https://api.lokalise.co/api/project/list?api_token=" + api_token)
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

				for _, v := range dat {
					switch vv := v.(type) {
					case []interface{}:
						for _, u := range vv {
							prj := u.(map[string]interface{})
							c_white := color.New(color.FgHiWhite)
							c_green := color.New(color.FgGreen)
							c_red := color.New(color.FgRed)
							c_cyan := color.New(color.FgCyan)

							c_white.Print(prj["id"])
							if prj["owner"] == "1" {
								c_green.Print(" (admin) ")
							} else {
								c_red.Print(" (contr) ")
							}
							c_cyan.Println(" ", prj["name"])
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

				if config_file == "" {
					config_file = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(config_file, &conf); err != nil {
					// do nothing if no config
				}

				if api_token == "" {
					api_token = conf.Token
				}
				if api_token == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				pid := c.Args().First();
				if (pid == "") {
					pid = conf.Project
				}
				if (pid == "") {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help export` for all options.", 5)
				}

				dest := c.String("dest")
				if dest == "" {
					dest = "."
				}

				filetype := c.String("type")
				if filetype == "" {
					return cli.NewExitError("ERROR: --type is required. Run `lokalise help export` for all options.", 5)
				}

				theurl := "api_token=" + api_token + `&id=` + pid + `&type=` + filetype

				langs := c.String("langs")
				if (langs != "") {
					langs = s2json(langs)
					theurl += "&langs=" + langs
				}

				use_original := c.String("use_original")
				if (use_original != "") {
					theurl += "&use_original=" + use_original
				}

				no_language_folders := c.String("no_language_folders")
				if (no_language_folders != "") {
					theurl += "&switch_no_language_folders=" + no_language_folders
				}

				filter := c.String("filter")
				if (filter != "") {
					filter = s2json(filter)
					theurl += "&filter=" + filter
				}

				triggers := c.String("triggers")
				if (triggers != "") {
					triggers = s2json(triggers)
					theurl += "&triggers=" + triggers
				}

				bundle_structure := c.String("bundle_structure")
				if (bundle_structure != "") {
					theurl += "&bundle_structure=" + bundle_structure
				}

				webhook_url := c.String("webhook_url")
				if (webhook_url != "") {
					theurl += "&webhook_url=" + webhook_url
				}

				export_all := c.String("export_all")
				if (export_all != "") {
					theurl += "&export_all=" + export_all
				}

				ota_plugin_bundle := c.String("ota_plugin_bundle")
				if (ota_plugin_bundle != "") {
					theurl += "&ota_plugin_bundle=" + ota_plugin_bundle
				}

				export_empty := c.String("export_empty")
				if (export_empty != "") {
					theurl += "&export_empty=" + export_empty
				}

				export_sort := c.String("export_sort")
				if (export_sort != "") {
					theurl += "&export_sort=" + export_sort
				}

				include_comments := c.String("include_comments")
				if (include_comments != "") {
					theurl += "&include_comments=" + include_comments
				}

				replace_breaks := c.String("replace_breaks")
				if (replace_breaks != "") {
					theurl += "&replace_breaks=" + replace_breaks
				}

				yaml_include_root := c.String("yaml_include_root")
				if (yaml_include_root != "") {
					theurl += "&yaml_include_root=" + yaml_include_root
				}

				json_unescaped_slashes := c.String("json_unescaped_slashes")
				if (json_unescaped_slashes != "") {
					theurl += "&json_unescaped_slashes=" + json_unescaped_slashes
				}

				include_pids := c.String("include_pids")
				if (include_pids != "") {
					include_pids = s2json(include_pids)
					theurl += "&include_pids=" + include_pids
				}

				tags := c.String("tags")
				if (tags != "") {
					tags = s2json(tags)
					theurl += "&tags=" + tags
				}

				unzip_to := c.String("unzip_to")

				sp := spinner.New(spinner.CharSets[9], 100 * time.Millisecond)
				sp.Start()
				fmt.Print("Requesting...")
				body := strings.NewReader(theurl)
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

				file_url := "https://s3-eu-west-1.amazonaws.com/lokalise-assets/" + file
				filename := strings.Split(file, "/")[4]

				sp.Stop()

				c_white := color.New(color.FgHiWhite)
				c_green := color.New(color.FgGreen)

				c_white.Println()
				c_white.Print("Remote ")
				c_green.Print(file_url + "... ")
				c_white.Println("OK")

				c_white.Print("Local ")
				c_green.Print(dest + "/" + filename + "... ")

				downloadFile(dest + "/" + filename, file_url)
				c_white.Println("OK")

				if (unzip_to != "") {
					files, err := Unzip(dest + "/" + filename, unzip_to)

					if err != nil {
						c_white.Println("Error unzipping files")
					} else {
						c_white.Print("Unzipped ")
						c_green.Print(strings.Join(files, ", ") + " ")
						c_white.Println("OK")
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

				if config_file == "" {
					config_file = "/etc/lokalise.cfg"
				}

				if _, err := toml.DecodeFile(config_file, &conf); err != nil {
					// do nothing if no config
				}

				if api_token == "" {
					api_token = conf.Token
				}
				if api_token == "" {
					return cli.NewExitError("ERROR: --token is required.  Run `lokalise help` for all options.", 5)
				}

				url := "https://api.lokalise.co/api/project/import"

				pid := c.Args().First();
				if (pid == "") {
					pid = conf.Project
				}
				if (pid == "") {
					return cli.NewExitError("ERROR: Project ID is required as first command option. Run `lokalise help import` for all options.", 5)
				}

				file := c.String("file")
				if (file == "") {
					return cli.NewExitError("ERROR: --file required.  Run `lokalise help import` for all options.", 5)
				}

				lang_iso := c.String("lang_iso")
				if lang_iso == "" {
					return cli.NewExitError("ERROR: --lang_iso is required. If you are using filemask in --file parameter, make sure escape it (e.g. \\*.json).  Run `lokalise help import` for all options. ", 5)
				}

				fill_empty := c.String("fill_empty")
				hidden := c.String("hidden")
				distinguish := c.String("distinguish")
				replace := c.String("replace")
				use_trans_mem := c.String("use_trans_mem")
				replace_breaks := c.String("replace_breaks")

				tags := c.String("tags")
				if (tags != "") {
					tags = s2json(tags)
				}

				extraParams := map[string]string{
					"id": pid,
					"api_token": api_token,
					"lang_iso": lang_iso,
					"fill_empty": fill_empty,
					"hidden": hidden,
					"distinguish": distinguish,
					"replace": replace,
					"tags": tags,
					"use_trans_mem": use_trans_mem,
					"replace_breaks": replace_breaks,
				}

				filemasks := strings.Split(file, ",")

				c_white := color.New(color.FgHiWhite)
				c_green := color.New(color.FgGreen)

				for _, mask := range filemasks {
					files, err := filepath.Glob(mask)

					if err != nil {
						log.Fatal(err)
					}

					for _, filename := range files {
						sp := spinner.New(spinner.CharSets[9], 100 * time.Millisecond)
						sp.Start()

						c_white.Print("Uploading ")
						c_white.Print(filename)
						c_white.Print("... ")

						request, err := newfileUploadRequest(url, extraParams, "file", filename)

						if err != nil {
							log.Fatal(err)
						}
						client := &http.Client{}
						resp, err := client.Do(request)
						if err != nil {
							log.Fatal(err)
						}
						sp.Stop()
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

							c_green.Print("Inserted ")
							c_white.Print(inserted)
							c_green.Print(", skipped ")
							c_white.Print(skipped)
							c_green.Print(", updated ")
							c_white.Print(updated)
							c_green.Println(" keys.")
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