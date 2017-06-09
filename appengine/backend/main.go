package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	"github.com/jlubawy/go-boilerpipe"
	"github.com/jlubawy/go-boilerpipe/backoff"
	"github.com/jlubawy/go-boilerpipe/extractor"

	"golang.org/x/net/context"
	"golang.org/x/net/publicsuffix"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.Handle("/api/extract", ApiHandler(ExtractArticle))
}

func ExtractArticle(ctx context.Context, resp http.ResponseWriter, req *http.Request) (error, int) {
	if req.Method != http.MethodGet {
		return fmt.Errorf("unexpected method '%s'", req.Method), http.StatusMethodNotAllowed
	}

	method := req.FormValue("method")
	if method == "" {
		method = http.MethodGet
	}
	contentType := req.FormValue("type")
	rawurl := req.FormValue("url")

	clientReq, err := http.NewRequest(method, rawurl, nil)
	if err != nil {
		return err, http.StatusBadRequest
	}

	client := urlfetch.Client(ctx)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	client.Jar = jar

	clientResp, err := backoff.Backoff(func() (*http.Response, error) {
		return client.Do(clientReq)
	})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	defer clientResp.Body.Close()

	document, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	r := bytes.NewReader(document)
	doc, err := boilerpipe.NewTextDocument(r, clientReq.URL)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	extractor.Article().Process(doc)
	doc.ContentType = contentType

	log.Infof(ctx, "contentType=%s", doc.ContentType)

	v := make(map[string]interface{})
	v["status"] = http.StatusOK
	v["results"] = doc
	if err := json.NewEncoder(resp).Encode(&v); err != nil {
		return err, http.StatusInternalServerError
	}

	return nil, http.StatusOK
}

func RunHandler(
	resp http.ResponseWriter, req *http.Request,
	handler func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (err error, statusCode int),
	errFn func(ctx context.Context, resp http.ResponseWriter, req *http.Request, err error, statusCode int),
) {
	ctx := appengine.NewContext(req)

	log.Infof(ctx, "%s %s", req.Method, req.RequestURI)

	defer func() {
		// Log any panics
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.Errorf(ctx, "handler panic: %s", err)
			} else {
				log.Errorf(ctx, "handler panic: %v", r)
			}
		}

		// Make sure to close the request body
		req.Body.Close()
	}()

	if err, statusCode := handler(ctx, resp, req); err != nil {
		log.Errorf(ctx, err.Error())
		resp.WriteHeader(statusCode)
		errFn(ctx, resp, req, err, statusCode)
	}
}

type ApiHandler func(ctx context.Context, resp http.ResponseWriter, req *http.Request) (error, int)

func (handler ApiHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Content-Type", "application/json")

	RunHandler(resp, req, handler, func(ctx context.Context, resp http.ResponseWriter, req *http.Request, err error, statusCode int) {
		v := make(map[string]interface{})
		v["status"] = statusCode
		v["message"] = err.Error()
		if err := json.NewEncoder(resp).Encode(&v); err != nil {
			log.Errorf(ctx, err.Error())
		}
	})
}
