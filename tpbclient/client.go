package tpbclient

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wael/music-streaming/models"
	"golang.org/x/net/html"
)

//Torrent represents a torrent magnet link from TPB
type Torrent struct {
	Name     string
	Link     string
	Seeders  int
	Leechers int
}

const query = `https://thepiratebay.org/search/%s/0/7/0`

//SearchRelease on TPB and returns the first page
// of results parsed into and array of Torrent structs
func SearchRelease(release models.Release) ([]Torrent, error) {
	if release.AlbumArtist == nil {
		return nil, errors.New("FindAndAddTPBTorrent: release has no artist associated")
	}
	return Search(release.AlbumArtist.Name + " " + release.Name)
}

//Search for searchTerm on TPB and returns the first page
// of results parsed into and array of Torrent structs
func Search(searchTerm string) ([]Torrent, error) {
	resp, err := http.Get(fmt.Sprintf(query, searchTerm))
	if err != nil {
		return nil, err
	}
	node, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	searchResults := searchTreeForSearchResult(node)
	if searchResults == nil {
		return nil, errors.New("could not find results")
	}
	node = searchResults.FirstChild
	var results []Torrent
	for node != nil {
		if node.Data == "tr" {
			torrent, err := extractDataFromTR(node)
			if err != nil {
				return nil, err
			}
			results = append(results, torrent)
		}
		node = node.NextSibling
	}
	return results, nil
}

func extractDataFromTR(tr *html.Node) (result Torrent, err error) {
	if tr == nil ||
		tr.FirstChild == nil ||
		tr.FirstChild.NextSibling == nil ||
		tr.FirstChild.NextSibling.NextSibling == nil ||
		tr.FirstChild.NextSibling.NextSibling.NextSibling == nil {
		err = errors.New("can't parse node")
		return
	}
	root := tr.FirstChild.NextSibling.NextSibling.NextSibling
	if root.FirstChild == nil ||
		root.FirstChild.NextSibling == nil ||
		root.FirstChild.NextSibling.FirstChild == nil ||
		root.FirstChild.NextSibling.FirstChild.NextSibling == nil ||
		root.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild == nil {
		err = errors.New("can't parse node")
		return
	}
	result.Name = root.FirstChild.NextSibling.FirstChild.NextSibling.FirstChild.Data
	if root.FirstChild.NextSibling.NextSibling == nil ||
		root.FirstChild.NextSibling.NextSibling.NextSibling == nil {
		err = errors.New("can't parse node")
		return
	}
	for _, attr := range root.FirstChild.NextSibling.NextSibling.NextSibling.Attr {
		if attr.Key == "href" {
			result.Link = attr.Val
		}
	}
	if root.NextSibling == nil ||
		root.NextSibling.NextSibling == nil ||
		root.NextSibling.NextSibling.FirstChild == nil {
		err = errors.New("can't parse node")
		return
	}
	result.Seeders, err = strconv.Atoi(root.NextSibling.NextSibling.FirstChild.Data)
	if err != nil {
		return
	}
	if root.NextSibling.NextSibling == nil ||
		root.NextSibling.NextSibling.NextSibling == nil ||
		root.NextSibling.NextSibling.NextSibling.NextSibling == nil ||
		root.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild == nil {
		err = errors.New("can't parse node")
		return
	}
	result.Leechers, err = strconv.Atoi(root.NextSibling.NextSibling.NextSibling.NextSibling.FirstChild.Data)
	return
}
func searchTreeForSearchResult(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}
	if node.Data == "tbody" {
		return node
	}
	if res := searchTreeForSearchResult(node.FirstChild); res != nil {
		return res
	}
	return searchTreeForSearchResult(node.NextSibling)
}
