package main
// #cgo CXXFLAGS: -std=c++11 -I/scratch/sammy/concur_caffe/include
// #cgo CXXFLAGS: -I/usr/local/cuda/include/  -I/scratch/sammy/opencv/include -O2 -fomit-frame-pointer -Wall
// #cgo LDFLAGS: -L/scratch/sammy/opencv/lib -lopencv_core  -lopencv_highgui -lopencv_imgproc -lopencv_imgcodecs
// #cgo LDFLAGS:  -lopencv_cudaarithm -lopencv_cudaimgproc -lopencv_cudawarping
// #cgo LDFLAGS: -L/scratch/sammy/concur_caffe/build/lib -lcaffe
// #cgo LDFLAGS: -L/usr/local/cuda/lib64/ -lcudart -lcublas -lcurand -lglog -lboost_system -lboost_thread
// #cgo LDFLAGS: -Wl,-rpath,/scratch/sammy/opencv/lib -Wl,-rpath,/scratch/sammy/concur_caffe/build/lib
// #include <stdlib.h>
// #include "classification.h"
import "C"
import "unsafe"

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var ctx *C.classifier_ctx

func classify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	buffer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cstr, err := C.classifier_classify(ctx, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer C.free(unsafe.Pointer(cstr))
	io.WriteString(w, C.GoString(cstr))
}

func main() {
	cmodel := C.CString(os.Args[1])
	ctrained := C.CString(os.Args[2])
	cmean := C.CString(os.Args[3])
	clabel := C.CString(os.Args[4])

	log.Println("Initializing Caffe classifiers")
	var err error
	ctx, err = C.classifier_initialize(cmodel, ctrained, cmean, clabel)
	if err != nil {
		log.Fatalln("could not initialize classifier:", err)
		return
	}
	defer C.classifier_destroy(ctx)

	log.Println("Adding REST endpoint /api/classify")
	http.HandleFunc("/api/classify", classify)
	log.Println("Starting server listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
