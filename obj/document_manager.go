package obj

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"code.sajari.com/docconv"
)

var (
	docMgr         *documentManager
	ErrUnsupported = errors.New("unsupported filetype")
)

type Config struct {
	DBPath        string `yaml:"db_path"`
	FileStorePath string `yaml:"file_store_path"`
}

func Setup(cfg Config) {
	var err error
	docMgr, err = NewDocumentManager(cfg.FileStorePath)
	if err != nil {
		panic(err)
	}

}

type documentManager struct {
	baseDir  string
	l        *sync.RWMutex
	saveChan chan *Document
}

func (d *documentManager) saveProcessor() {
	for {
		err := d.save(<-d.saveChan)
		if err != nil {
			fmt.Println("SAVE:", err.Error())
		}
	}
}

func (d documentManager) txtPath() string {
	return path.Join(d.baseDir, "text")
}
func (d documentManager) otherPath() string { return path.Join(d.baseDir, "other") }

func (d documentManager) pdfPath() string {
	return path.Join(d.baseDir, "pdf")
}

func NewDocumentManager(baseDir string) (d *documentManager, err error) {
	d = &documentManager{baseDir: baseDir, l: &sync.RWMutex{}, saveChan: make(chan *Document, 512)}
	if err = os.Mkdir(baseDir, 0777); err != nil && !os.IsExist(err) {
		return
	}
	if err = os.Mkdir(d.txtPath(), os.ModePerm); err != nil && !os.IsExist(err) {
		return
	}
	if err = os.Mkdir(d.pdfPath(), 0777); err != nil && !os.IsExist(err) {
		return
	}
	err = nil
	go d.saveProcessor()
	return
}

func (d *documentManager) save(doc *Document) error {
	b, err := doc.retrieve()
	if err != nil {
		return err
	}
	f, err := os.Create(d.genPath(doc))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)

	if err != nil {
		return err
	}

	return d.saveText(doc)
}

func (d *documentManager) remove(doc *Document) error {

	return nil
}

func (d *documentManager) load(doc *Document) ([]byte, error) {
	if b, err := os.ReadFile(d.genPath(doc)); err == nil {
		return b, err
	}
	err := d.save(doc)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(d.genPath(doc))
}

func (d *documentManager) loadFile(doc *Document) (*os.File, error) {
	return os.Open(d.genPath(doc))
}

func (d *documentManager) loadText(doc *Document) ([]byte, error) {
	return os.ReadFile(path.Join(d.txtPath(), doc.Id, ".txt"))
}

func (d *documentManager) genPath(doc *Document) string {
	switch doc.ContentType {
	case "text/plain":
		return path.Join(d.txtPath(), doc.Id+".txt")
	case "text/html":
		return path.Join(d.txtPath(), doc.Id+".html")
	case "text/xml":
		return path.Join(d.txtPath(), doc.Id+".xml")
	case "application/pdf":
		return path.Join(d.pdfPath(), doc.Id+".pdf")
	default:
		return path.Join(d.otherPath(), doc.Id)
	}
}

func (d *documentManager) saveText(doc *Document) error {
	if strings.Contains(doc.ContentType, "text/") {
		return nil
	}
	if doc.ContentType != "application/pdf" {
		return ErrUnsupported
	}
	_, err := d.loadFile(doc)
	if err != nil {
		return err
	}
	res, err := docconv.ConvertPath(d.genPath(doc))
	if err != nil {
		return err
	}
	pth := path.Join(d.txtPath(), doc.Id+".txt")
	f, err := os.Create(pth)
	if err != nil {
		return err
	}
	_, err = f.WriteString(res.Body)

	return err
}
