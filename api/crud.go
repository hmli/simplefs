package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/hmli/simplefs/core"
	"github.com/hmli/simplefs/utils"
	"strconv"
)

type Server struct {
	Mux    *http.ServeMux
	Port   int
	Volume *core.Volume
}

// TODO 把 fmt.Print 改成 log
func NewServer(port int, dir string) *Server {
	v, err := core.NewVolume(1, dir)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	return &Server{
		Mux:    http.NewServeMux(),
		Port:   port,
		Volume: v,
	}
}

func (s *Server) Run() {
	s.Mux.HandleFunc("/img", s.FileHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Mux)
	if err != nil {
		panic(err)
	}
}

func (s *Server) FileHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println(r.Method)
	switch r.Method {
	case "GET": // TODO cache with E-Tag
		r.ParseForm()
		id := r.Form.Get("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprint(w, "Wrong format id: ", id)
			return
		}
		data, ext, err := s.Volume.GetFile(uint64(idInt))
		if err != nil {
			fmt.Fprint(w, "No file")
			return
		}
		fmt.Println("Ext: ", ext, "lendata: ", len(data))
		w.Header().Set("Content-Type", ContentType(ext))
		fmt.Fprint(w, string(data))
		return
	case "POST":
		r.ParseMultipartForm(32 << 20)
		fmt.Println(r.MultipartForm)
		fmt.Println(r.Form)
		file, header, err := r.FormFile("file")
		fmt.Printf("file: %+v, %+v, err %s", file, header, err)
		if err != nil {
			fmt.Println(err)
			fmt.Fprint(w, "Upload fail")
			return
		}
		filename := header.Filename
		fmt.Println("Filename:", filename)
		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprint(w, "File content err")
			return
		}
		id, err := s.Volume.NewFile(data, filename)
		if err != nil {
			fmt.Fprint(w, "File storing err")
			return
		}
		fmt.Fprint(w, id)
	case "DELETE":
		r.ParseForm()
		id := r.Form.Get("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			fmt.Fprint(w, "Wrong format id: ", id)
			return
		}
		err = s.Volume.DelNeedle(uint64(idInt))
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		fmt.Fprint(w, id)
	default:
		fmt.Fprint(w, "Invalid method")
	}
}

func ContentType(ext string) (ctype string) {
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "pdf":
		return "application/pdf"
	case "png":
		return "image/png"
	case "json":
		return "application/json"
	case "js":
		return "applicaton/javascript"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-sream"
	}
}
