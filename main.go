package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Strubbl/wallabago"
	"golang.org/x/net/html"
)

var debug = flag.Bool("d", false, "get debug output (implies verbose mode)")
var verbose = flag.Bool("verbose", false, "verbose mode")
var configJSON = flag.String("config", "config.json", "file name of config JSON file")
var pocketFile = flag.String("pocketfile", "ril_export.html", "file name of the HTML export of your Pocket entries, downloaded from https://getpocket.com/export")
var tagArchivedEntries = flag.Int("archives", 0, "Set to 1 if you want to tag archived wallabag items also?")

func handleFlags() {
	flag.Parse()
	if *debug && len(flag.Args()) > 0 {
		log.Printf("handleFlags: non-flag args=%v", strings.Join(flag.Args(), " "))
	}
	// test verbose before debug because debug implies verbose
	if *verbose && !*debug {
		log.Printf("verbose mode")
	}
	if *debug {
		log.Printf("handleFlags: debug mode")
		// debug implies verbose
		*verbose = true
	}
}

type PocketEntry struct {
	Tags          []string
	Url           string
	RedirectedUrl string
	WallabagItem  *wallabago.Item
}

func main() {
	log.SetOutput(os.Stdout)
	handleFlags()
	// check for config
	if *verbose {
		log.Println("reading config", *configJSON)
	}
	err := wallabago.ReadConfig(*configJSON)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	// Read Pocket entries, without knowledge of archival status
	pocketEntries := getPocketEntries(*pocketFile)
	log.Printf("Read %v pocket items from file\n", len(pocketEntries))
	log.Println()

	// Get Wallabag entries from cache
	var wallabagItems []wallabago.Item
	_, e1 := os.Stat(".wallabag_bin_items")
	if !os.IsNotExist(e1) {
		wallabagItemBinFile, err := os.Open(".wallabag_bin_items")
		check(err)
		defer wallabagItemBinFile.Close()
		decoder := gob.NewDecoder(wallabagItemBinFile)
		err = decoder.Decode(&wallabagItems)
		if err != nil {
			log.Fatal("decode error 1:", err)
		}
		log.Printf("Read %v wallabag items from file\n", len(wallabagItems))
	} else {
		// Cache failed, retrieve all and populate cache
		wallabagItems, e2 := getWallabagItems()
		check(e2)
		wallabagItemFile, e3 := os.Create(".wallabag_items")
		check(e3)
		defer wallabagItemFile.Close()
		wallabagItemBinFile, e4 := os.Create(".wallabag_bin_items")
		check(e4)
		enc := gob.NewEncoder(wallabagItemBinFile)
		defer wallabagItemBinFile.Close()
		for _, item := range wallabagItems {
			output := item.URL + "\n\n"
			_, e5 := wallabagItemFile.WriteString(output)
			check(e5)
		}
		// Cache for next time
		e6 := enc.Encode(wallabagItems)
		if e6 != nil {
			log.Fatal("encode error:", e6)
		}
		log.Printf("Wrote %v wallabag items to file\n", len(wallabagItems))
	}
	log.Println()

	matchedFile, err := os.Create(".matchedPocketEntries")
	check(err)
	defer matchedFile.Close()

	// For each pocket entry, try to find a wallabag entry to tag
	log.Println("Basic URL matching...")
	matchCount := 0
	var matchedWallabagIndices []int
	for i, pocketEntry := range pocketEntries {
		foundMatch := false
		canonizedPocketUrl := canonicalizeUrl(pocketEntry.Url)
		for j, item := range wallabagItems {
			canonizedItemUrl := canonicalizeUrl(item.URL)
			if canonizedPocketUrl == canonizedItemUrl {
				matchCount++
				foundMatch = true
				matchedWallabagIndices = append(matchedWallabagIndices, j)
				pocketEntries[i].WallabagItem = &item

				output := pocketEntry.Url + "\nmatched\n" + item.URL + "\n\n"
				_, err := matchedFile.WriteString(output)
				check(err)

				break
			}
		}
		if foundMatch {
			fmt.Printf("✓")
		} else {
			fmt.Printf("x")
		}
	}
	fmt.Println()

	// Make a second pass with URL redirection (slow) for unmatched URLs
	log.Println("Now see if remaining URLs to match can be fixed via redirection")
	ignoreCount := 0
	redirectedMatchCount := 0
	unmatchedFile, err := os.Create(".unmatchedPocketEntries")
	check(err)
	defer unmatchedFile.Close()
	for i, pocketEntry := range pocketEntries {
		if pocketEntry.WallabagItem != nil {
			fmt.Printf(".")
			ignoreCount++
			continue
		}
		fmt.Printf("\nChecking redirect for unmatched Pocket URL: %s\n", pocketEntry.Url)
		foundMatch := false
		pocketEntry.RedirectedUrl = getRedirectedUrl(pocketEntry.Url)
		canonizedRedirectedPocketUrl := canonicalizeUrl(pocketEntry.RedirectedUrl)
		for j, item := range wallabagItems {
			canonizedItemUrl := canonicalizeUrl(item.URL)
			if canonizedRedirectedPocketUrl == canonizedItemUrl {
				foundMatch = true
				matchCount++
				redirectedMatchCount++
				pocketEntries[i].WallabagItem = &item
				output := pocketEntry.Url + "\nmatched\n" + item.URL + "\n\n"
				_, err := matchedFile.WriteString(output)
				check(err)
				matchedWallabagIndices = append(matchedWallabagIndices, j)
				break
			}
		}
		if !foundMatch {
			output := pocketEntry.Url + "\n\n"
			_, err := unmatchedFile.WriteString(output)
			check(err)
		}
	}
	fmt.Println()
	log.Printf("Matched %v pocket entries with URL redirection pass\n", redirectedMatchCount)
	log.Printf("Ignored %v pocket entries in URL redirection pass\n", ignoreCount)

	// Output which Wallabag items didn't get matched to help debugging
	var unmatchedWallabagItems []string
	unmatchedWallabagFile, err := os.Create(".unmatchedWallabagEntries")
	check(err)
	defer unmatchedWallabagFile.Close()
	for i, item := range wallabagItems {
		wasMatched := false
		for _, index := range matchedWallabagIndices {
			if i == index {
				wasMatched = true
				break
			}
		}
		if !wasMatched {
			unmatchedWallabagItems = append(unmatchedWallabagItems, item.URL)
			output := item.URL + "\n\n"
			_, err := unmatchedWallabagFile.WriteString(output)
			check(err)
		}
	}

	log.Printf("Found %v pocket entries\n", len(pocketEntries))
	log.Printf("Found %v wallabag entries\n", len(wallabagItems))
	log.Printf("Found %v matches\n", matchCount)
	log.Printf("%v unmatched pocket entries (see .unmatchedPocketEntries)\n", len(pocketEntries)-matchCount)
	log.Printf("%v unmatched wallabag entries (see .unmatchedWallabagEntries)\n", len(unmatchedWallabagItems))

	log.Println()
	log.Println("Setting Wallabag tags")
	for _, pocketEntry := range pocketEntries {
		item := pocketEntry.WallabagItem
		if item == nil {
			continue
		}
		err := wallabago.AddEntryTags(item.ID, pocketEntry.Tags...)
		if err != nil {
			log.Println("AddEntryTags failure: ", err)
			log.Println(pocketEntry)
			log.Println(item)
			os.Exit(1)
		}
		tagList := strings.Join(pocketEntry.Tags, ",")
		log.Printf("Tagged Wallabag entry ID %v with %s\nURL: %s", item.ID, tagList, item.URL)
	}
}

func getWallabagItems() ([]wallabago.Item, error) {
	log.Println("Getting Wallabag entries")
	archive := *tagArchivedEntries
	starred := -1
	sort := ""
	order := ""
	page := -1
	perPage := -1
	tags := ""
	e, err := wallabago.GetEntries(
		wallabago.APICall,
		archive,
		starred,
		sort,
		order,
		page,
		perPage,
		tags)
	if err != nil {
		log.Println("GetAllEntries: first GetEntries call failed", err)
		return nil, err
	}
	allEntries := e.Embedded.Items
	if e.Total > len(allEntries) {
		fmt.Printf(".")
		secondPage := e.Page + 1
		perPage = e.Limit
		pages := e.Pages
		for i := secondPage; i <= pages; i++ {
			e, err := wallabago.GetEntries(
				wallabago.APICall,
				archive,
				starred,
				sort,
				order,
				i,
				perPage,
				tags)
			if err != nil {
				log.Printf("GetAllEntries: GetEntries for page %d failed: %v", i, err)
				return nil, err
			}
			tmpAllEntries := e.Embedded.Items
			allEntries = append(allEntries, tmpAllEntries...)
		}
	}
	// For easier debugging, strip the content from the items
	for i, _ := range allEntries {
		allEntries[i].Content = "Content hidden for easier debugging"
	}
	return allEntries, err
}

func getPocketEntries(filename string) []PocketEntry {
	log.Printf("Reading %s...", filename)
	log.Println()
	var entries []PocketEntry

	f, err := os.Open(filename)
	check(err)
	defer f.Close()
	tokenizer := html.NewTokenizer(f)
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			return entries
		}

		tagName, _ := tokenizer.TagName()
		if len(tagName) == 1 && tagName[0] == 'a' && tt == html.StartTagToken {
			var entry = new(PocketEntry)

			for {
				keyBytes, valBytes, more := tokenizer.TagAttr()
				key := string(keyBytes)
				val := string(valBytes)
				if key == "href" {
					entry.Url = val
				}
				if key == "tags" {
					entry.Tags = strings.Split(val, ",")
				}
				if !more {
					break
				}
			}
			entries = append(entries, *entry)
		}

	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRedirectedUrl(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("")
		log.Printf("error in http.Get => %v\n", err.Error())
		return ""
	} else {
		return resp.Request.URL.String()
	}
}

func canonicalizeUrl(url string) string {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	url = strings.Split(url, "?")[0]
	url = strings.Split(url, "#")[0]
	url = strings.TrimSuffix(url, "/")
	return url
}
