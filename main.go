package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const (
	URL             = "https://www.amazon.in"
	PATH            = "Savedimg"
	DownloadChannel = 1
)

func varInit() map[string]bool {
	imagesUrl := make(map[string]bool)
	return imagesUrl
}

func urlToHtml(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	res, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func urlLink(url string) (map[string]struct{}, error) {
	resp, err := urlToHtml(url)
	if err != nil {
		return nil, err
	}

	imagesUrl := make(map[string]struct{})
	resp.Find("*").Each(func(index int, item *goquery.Selection) {
		tag := item.Find("img")
		link, _ := tag.Attr("src") //link ,bool
		if link != "" {
			imagesUrl[link] = struct{}{}
		}
	})
	return imagesUrl, nil
}

func downloader(imageUrl map[string]struct{}) error {
	if _, err := os.Stat(PATH); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(PATH, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	var sem chan struct{}
	var wg sync.WaitGroup
	sem = make(chan struct{}, DownloadChannel)
	defer close(sem)
	errs := make(chan error)
	for value := range imageUrl { //val,bool
		wg.Add(1)
		select {
		case sem <- struct{}{}:

		case x := <-errs:
			return x

		}

		go func(val string) {
			err := downlaodImage(val)
			if err != nil {
				errs <- err
			}

			defer wg.Done()
			defer func() {

				<-sem
			}()

		}(value)
	}
	wg.Wait()

	return nil
}

func downlaodImage(ImagesUrl string) error {

	var Addurl string
	if ImagesUrl[:4] != "http" {
		Addurl = "http:" + ImagesUrl
	} else {
		Addurl = ImagesUrl
	}
	parts := strings.Split(Addurl, "/")

	name := parts[len(parts)-1]

	resp, err := http.Get(Addurl)
	if err != nil {
		return errors.New("unable to fetch url")
	}

	file, err := os.Create(string(PATH + "/" + name))
	if err != nil {
		return errors.New("unable to create file ")

	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return errors.New("unable to copy")
	}
	fmt.Printf("Saving %s \n", PATH+"/"+name)
	return nil
}

func main() {
	server()
	imageurl, err := urlLink(URL)
	if err != nil {
		log.Fatal(err)
	}

	err = downloader(imageurl)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Downloaded completed")
}

func server() {
	http.HandleFunc("/", RootHandler)

	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal(err)
	}

}
func RootHandler(res http.ResponseWriter, req *http.Request) {
	file, _ := ioutil.ReadFile("img.html")
	res.Write(file)
}
