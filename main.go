package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/linksmart/historical-datastore/data"
	"github.com/linksmart/historical-datastore/registry"
	"github.com/linksmart/service-catalog/v3/utils"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("usage: %s sourceUrl destUrl", filepath.Base(os.Args[0]))
		return
	}

	sourceUrl := strings.TrimSuffix(os.Args[1], "/")
	dstUrl := strings.TrimSuffix(os.Args[2], "/")

	srcRegEndpoint := sourceUrl + "/registry"
	srcRegClient, err := registry.NewRemoteClient(srcRegEndpoint, nil)
	if err != nil {
		fmt.Printf("Error creating the src registry client:%v", err)
		return
	}

	dstRegEndpoint := dstUrl + "/registry"
	dstRegClient, err := registry.NewRemoteClient(dstRegEndpoint, nil)
	if err != nil {
		fmt.Printf("Error creating the dst registry client:%v", err)
		return
	}

	/*
		srcDataClient, err := data.NewRemoteClient(srcDataEndpoint, nil)

		if err != nil {
			log.Panic("Error creating the src data client")
		}
	*/

	dstDataEndpoint := dstUrl + "/data"
	dstDataClient, err := data.NewRemoteClient(dstDataEndpoint, nil)
	if err != nil {
		fmt.Printf("Error creating the dst data client:%v", err)
		return
	}

	perPage := 100
	page := 1

	condition := true

	for condition {
		streamList, err := srcRegClient.GetMany(page, perPage)
		if err != nil {
			fmt.Printf("Error fetching the registry entries:%v", err)
			return
		}

		left := streamList.Total - page*perPage
		condition = left > 0
		page++

		for _, stream := range streamList.Streams {
			fmt.Printf("copying the stream %s", stream.Name)
			_, err := dstRegClient.Add(&stream)
			if err != nil {
				if !strings.HasPrefix(err.Error(), strconv.Itoa(http.StatusConflict)) {
					fmt.Printf("Error creating the registry entry for %s.. Skipping:%s", stream.Name, err)
					continue //to next stream
				}
			}

			err = copyData(sourceUrl, dstDataClient, stream)
			if err != nil {
				fmt.Printf("Error copying the data entry:%v", err)
				continue //to next stream
			}
		}
	}

}

func queryLink(path string) (*data.RecordSet, error) {

	res, err := utils.HTTPRequest("GET",
		path,
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body of response: %v", err.Error())
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v: %v", res.StatusCode, string(body))
	}

	var rs data.RecordSet
	err = json.Unmarshal(body, &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil

}

func copyData(srcUrl string, dst *data.RemoteClient, stream registry.DataStream) error {
	condition := true
	path := fmt.Sprintf("%s/data/%s", srcUrl, stream.Name)

	for condition {
		recordSet, err := queryLink(path)
		if err != nil {
			return fmt.Errorf("error retrieving the data for %s:%v", stream.Name, err)
		}
		srcPack := recordSet.Data
		condition = recordSet.NextLink != ""
		path = srcUrl + recordSet.NextLink
		b, _ := json.Marshal(srcPack)

		err = dst.Submit(b, "application/senml+json", stream.Name)
		if err != nil {
			return fmt.Errorf("error submitting the data for %s:%v", stream.Name, err)
		}
	}
	return nil
}
