package gopirate

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

//Torrent represents a torrent magnet link from TPB
type Torrent struct {
	Name     string
	Link     string
	Size     int64
	Seeders  int
	Leechers int
}

type direction int

const (
	query           = `https://thepiratebay.org/search/%s/0/7/0`
	fC    direction = iota
	nS
)

//ErrNoResults no results found
var ErrNoResults = errors.New("could not find results")

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
		return nil, ErrNoResults
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

func extractDataFromTR(tr *html.Node) (Torrent, error) {
	var result Torrent
	root, err := extractPath(tr, []direction{fC, nS, nS, nS})
	if err != nil {
		return result, fmt.Errorf("can't find root node: %v", err)
	}

	nameNode, err := extractPath(root, []direction{fC, nS, fC, nS, fC})
	if err != nil {
		return result, fmt.Errorf("can't find name node: %v", err)
	}
	result.Name = nameNode.Data

	linkNode, err := extractPath(root, []direction{fC, nS, nS, nS})
	if err != nil {
		return result, fmt.Errorf("can't find link node: %v", err)
	}

	for _, attr := range linkNode.Attr {
		if attr.Key == "href" {
			result.Link = attr.Val
		}
	}

	seedersNode, err := extractPath(root, []direction{nS, nS, fC})
	if err != nil {
		return result, fmt.Errorf("can't find seeders node: %v", err)
	}
	result.Seeders, err = strconv.Atoi(seedersNode.Data)
	if err != nil {
		return result, err
	}

	leechersNode, err := extractPath(root, []direction{nS, nS, nS, nS, fC})
	if err != nil {
		return result, fmt.Errorf("can't find leechers node: %v", err)
	}
	result.Leechers, err = strconv.Atoi(leechersNode.Data)
	infoNode, err := extractPath(root, []direction{fC, nS, nS, nS, nS, nS, nS, nS, fC})
	if err != nil {
		return result, fmt.Errorf("can't find info node: %v", err)
	}
	regMatch := regexp.MustCompile(`Size (?P<size>\d+\.\d+) (?P<sizetype>[K|M|G])iB`).FindStringSubmatch(infoNode.Data)
	if len(regMatch) > 2 {
		szFlt, _ := strconv.ParseFloat(regMatch[1], 32)
		result.Size = int64(szFlt * float64(map[string]int64{"K": 1024, "M": 1048576, "G": 1073741824}[regMatch[2]]))
	}
	return result, nil
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

func searchNodeForType(node *html.Node, data string) ([]direction, *html.Node) {
	if node == nil {
		return nil, nil
	}
	if node.Data == data {
		return nil, node
	}
	if path, resNode := searchNodeForType(node.FirstChild, data); resNode != nil {
		return append([]direction{fC}, path...), resNode
	}
	if path, resNode := searchNodeForType(node.NextSibling, data); resNode != nil {
		return append([]direction{nS}, path...), resNode
	}
	return nil, nil
}

func extractPath(node *html.Node, path []direction) (*html.Node, error) {
	if node == nil {
		return nil, errors.New("can't parse node")
	}
	if len(path) < 1 {
		return node, nil
	}
	switch path[0] {
	case fC:
		return extractPath(node.FirstChild, path[1:])
	case nS:
		return extractPath(node.NextSibling, path[1:])
	default:
		return nil, fmt.Errorf("unknown direction %v", path[0])
	}
}
