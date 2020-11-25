# wallabag_import_pocket_tags

A Go script to copy tags from an export of Pocket articles to your Wallabag articles. This program assumes you have already used the standard wallabag import from Pocket, which does not copy tags.

To get the URLs to match as well as possible, it performs basic URL manipulation like removing query strings, http vs https, and `/` suffixes. It also follows the Pocket URLs to check for HTTP redirects that may have been created since the article was saved to Pocket (Pocket seems to only save the original URL you added).

It's not perfect, but for me it achieved a ~95% success rate. The articles that failed to match were largely ones whose old URLs no longer work at all. All the unmatched entries (from Pocket and from Wallabag) are listed in debug output files - `.unmatchedPocketEntries` and `.unmatchedWallabagEntries`.

Usage:
1. Use [Wallabag's Pocket import](https://doc.wallabag.org/en/user/import/pocket)
2. Download your Pocket articles via [Pocket's export tool](https://getpocket.com/export)
3. Create a new API client on your Wallabag instance, for example at https://app.wallabag.it/developer/client/create
4. Copy those credentials along with your Wallabag server URL & login credentials into a config.json file, as in this example: https://github.com/Strubbl/wallabago/blob/master/example/config.json
5. go run main.go

Notes:
- Your Pocket export is named `ril_export.html`. If not, use the `-pocketfile=` parameter.
- You don't want to bother copying tags for all the articles you've already archived in Wallabag. If you want to tag ALL your Wallabag entries, use `-archives=2`

## With thanks to

[Wallabago](https://github.com/Strubbl/wallabago) - Go wrapper for the Wallabag API
