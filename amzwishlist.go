// Package amzwishlist provides a simple library to scrape contents
//from an Amazon wishlist for further processing
package amzwishlist

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/bjarneh/latinx"
	"gopkg.in/xmlpath.v2"

	"code.google.com/p/go.net/html"
)

// Wish represents a single wish in an Amazon wishlist
type Wish struct {
	Title, Price        string
	Requested, Received int
}

/*
ScrapeWishlist takes the wishlist ID (like "A9MBI81ZSO7Q") as
an argument and scrapes the wishlist items from the compact view of the
wishlists website.
*/
func ScrapeWishlist(wishlist string) []Wish {
	url := fmt.Sprintf("http://www.amazon.de/registry/wishlist/%s/?reveal=unpurchased&filter=all&sort=date-added&layout=compact", wishlist)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		panic(readErr)
	}

	good, converr := latinx.Decode(latinx.ISO_8859_15, body)
	if converr != nil {
		panic(converr)
	}

	reader := strings.NewReader(sanitizeHTML(string(good)))
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)
	if xmlerr != nil {
		panic(xmlerr)
	}

	pathTable := xmlpath.MustCompile("//*[contains(@class,'g-compact-items')]/tbody/tr[preceding::tr]")
	iterator := pathTable.Iter(xmlroot)
	wishes := []Wish{}

	for iterator.Next() {
		line := iterator.Node()

		title := getXpathString("td[@class='g-title']/a[@class='a-link-normal']", line)
		price := getXpathString("td[contains(@class,'g-price')]/span", line)
		requested := getXpathInt("td[@class='g-requested']/span", line)
		gotten := getXpathInt("td[@class='g-received']/span", line)
		result := Wish{title, price, requested, gotten}
		wishes = append(wishes, result)
	}

	return wishes
}

func getXpathString(xpath string, node *xmlpath.Node) string {
	path := xmlpath.MustCompile(xpath)
	if res, ok := path.String(node); ok {
		res = strings.Replace(res, "\u200b", "", -1) // WTF? Damn non-printable chars
		return strings.TrimSpace(res)
	}
	panic(fmt.Sprintf("No string found for %s", xpath))
}

func getXpathInt(xpath string, node *xmlpath.Node) int {
	str := getXpathString(xpath, node)
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return i
}

func sanitizeHTML(rawHTML string) string {
	reader := strings.NewReader(rawHTML)
	root, err := html.Parse(reader)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	html.Render(&b, root)
	fixedHTML := b.String()

	return fixedHTML
}
