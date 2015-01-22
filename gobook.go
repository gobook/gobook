package gobook

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobook/blackfriday"
)

func makeSummary(summary, linkPrefix string) (string, error) {
	f, err := os.Open(summary)
	if err != nil {
		return "", err
	}
	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	ct := strings.Replace(string(bs), "# Summary", "", -1)

	return MarkdownToHtml(ct, blackfriday.HtmlRendererParameters{
		AbsolutePrefix: linkPrefix,
	}), nil
}

func loadTemplate(tmplPath string) (*template.Template, error) {
	bs, err := ioutil.ReadFile(tmplPath)
	if err != nil {
		return nil, err
	}
	return template.New("frame").Parse(string(bs))
}

func makePage(dstPath, srcPath, linkPrefix string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	data := MarkdownToHtml(string(bs), blackfriday.HtmlRendererParameters{
		AbsolutePrefix: linkPrefix,
	})

	d, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = d.Write([]byte(data))
	return err
}

type Book struct {
	Name   string `json:"name"`
	Author string `json:"author"`
	Lang   string `json:"lang"`
	Desc   string `json:"desc"`
}

func loadBook(cfgPath string) (*Book, error) {
	bs, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var book Book
	err = json.Unmarshal(bs, &book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

func MakeBook(dstDir, srcDir string) error {
	book, err := loadBook(filepath.Join(srcDir, "book.json"))
	if err != nil {
		return err
	}

	os.RemoveAll(dstDir)
	os.MkdirAll(dstDir, os.ModePerm)

	mdSummary := filepath.Join(srcDir, "SUMMARY.md")
	summary, err := makeSummary(mdSummary, "./")
	if err != nil {
		return err
	}

	summary = strings.Replace(summary, `README.md"`, `index.html"`, -1)
	summary = strings.Replace(summary, `.md"`, `.html"`, -1)

	tmpl, err := loadTemplate("./themes/gitbook/templates/frame.html")
	if err != nil {
		return err
	}

	mdReadme := filepath.Join(srcDir, "README.md")
	readme, err := ioutil.ReadFile(mdReadme)
	if err != nil {
		return err
	}

	readmeHTML := MarkdownToHtml(string(readme), blackfriday.HtmlRendererParameters{
		AbsolutePrefix: "./",
	})

	var levelPath = "./"

	bf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(bf, "frame", map[string]interface{}{
		"summary":   template.HTML(summary),
		"content":   template.HTML(readmeHTML),
		"levelPath": levelPath,
		"book":      book,
		"title":     "首页",
	})
	if err != nil {
		return err
	}
	var output = bf.Bytes()

	f, err := os.Create(filepath.Join(dstDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(output))

	err = filepath.Walk(srcDir, func(path string, fi os.FileInfo, err error) error {
		if path == mdReadme || path == mdSummary {
			return nil
		}
		rPath, _ := filepath.Rel(srcDir, path)
		if strings.HasPrefix(rPath, "_book") || strings.HasPrefix(rPath, "themes") {
			return nil
		}
		if fi.IsDir() {
			os.MkdirAll(filepath.Join(dstDir, rPath), os.ModePerm)
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		page, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		level := strings.Count(rPath, string(filepath.Separator))
		var levelPath = ""
		for i := 0; i < level; i++ {
			levelPath = levelPath + "../"
		}
		if levelPath == "" {
			levelPath = "./"
		}

		pageHTML := MarkdownToHtml(string(page), blackfriday.HtmlRendererParameters{
			AbsolutePrefix: levelPath,
		})

		mdSummary := filepath.Join(srcDir, "SUMMARY.md")
		summary, err := makeSummary(mdSummary, levelPath)
		if err != nil {
			return err
		}

		summary = strings.Replace(summary, `README.md"`, `index.html"`, -1)
		summary = strings.Replace(summary, `.md"`, `.html"`, -1)

		bf := new(bytes.Buffer)
		err = tmpl.Execute(bf, map[string]interface{}{
			"summary":   template.HTML(summary),
			"content":   template.HTML(pageHTML),
			"levelPath": levelPath,
			"book":      book,
			"title":     "",
		})
		if err != nil {
			return err
		}
		var output = bf.Bytes()

		newpath := filepath.Join(dstDir, rPath)
		newpath = strings.Replace(newpath, "README.md", "index.html", -1)
		newpath = strings.Replace(newpath, ".md", ".html", -1)
		f, err := os.Create(newpath)
		if err != nil {
			return err
		}
		f.Write([]byte(output))
		f.Close()
		return nil
	})

	err = CopyDir("./themes/gitbook/asserts",
		filepath.Join(dstDir, "gitbook"))

	return err
}
