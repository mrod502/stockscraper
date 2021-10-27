package obj

import (
	"os"
	"path"
	"sync"

	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

var (
	docMgr *documentManager
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
		d.save(<-d.saveChan)
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
	if err = os.Mkdir(baseDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return
	}
	if err = os.Mkdir(d.txtPath(), os.ModePerm); err != nil && !os.IsExist(err) {
		return
	}
	if err = os.Mkdir(d.pdfPath(), 0600); err != nil && !os.IsExist(err) {
		return
	}
	go d.saveProcessor()
	return
}

func (d *documentManager) save(doc *Document) error {
	b, err := doc.retrieve()
	if err != nil {
		return err
	}
	f, err := os.Create(path.Join())
	if err != nil {
		return err
	}
	_, err = f.Write(b)

	if err != nil {
		return err
	}
	return nil
}

func (d *documentManager) remove(doc *Document) error {

	return nil
}

func (d *documentManager) load(doc *Document) ([]byte, error) {
	return os.ReadFile(d.genPath(doc))
}

func (d *documentManager) loadText(doc *Document) ([]byte, error) {
	return os.ReadFile(path.Join(d.txtPath(), string(doc.Id[:]), ".txt"))
}

func extractText(r *model.PdfReader) ([]byte, error) {
	var b []byte = make([]byte, 0)

	numPages, err := r.GetNumPages()
	if err != nil {
		return nil, err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := r.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return nil, err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return nil, err
		}
		b = append(b, []byte(text+"\n")...)

	}

	return b, nil
}

func (d *documentManager) genPath(doc *Document) string {
	switch doc.ContentType {
	case "text/plain":
		return path.Join(d.txtPath(), string(doc.Id[:])+".txt")
	case "text/html":
		return path.Join(d.txtPath(), string(doc.Id[:])+".html")
	case "text/xml":
		return path.Join(d.txtPath(), string(doc.Id[:])+".xml")
	case "application/pdf":
		return path.Join(d.pdfPath(), string(doc.Id[:])+".pdf")
	default:
		return path.Join(d.otherPath(), string(doc.Id[:]))
	}
}

func (d *documentManager) saveText(doc *Document) error {
	var b = make([]byte, 0)
	f, err := os.Open(d.genPath(doc))
	if err != nil {
		return err
	}
	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return err
		}
		b = append(b, []byte(text)...)

	}
	ft, err := os.Create(path.Join(d.txtPath(), string(doc.Id[:])+".txt"))
	if err != nil {
		return err
	}
	_, err = ft.Write(b)

	return err
}
