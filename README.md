# mediathek-scraper

Command line utility to search for and download VODs from the ARD mediathek.  
youtube-dl needs to be installed for downloading to work.

## Flags:  

### Mandatory:
      -search string
            the term to search for
### Optional:
      -download
            download the search results
      -path string
            the location to save the downloaded files (default is in the current directory)
      -workers int
            the maximum number of parallel downloads (default 1)
      -maxduration int
            the maximum duration (in seconds) for a VOD to be considured (default -1)
      -minduration int
            the minimum duration (in seconds) for a VOD to be considured (default -1)
      -regex string
            regular expression that needs to be matched by the title of a VOD to be considured


### Example usage:
    $ mediathek-scraper -search "Tatort" -minduration 1200 -regex "^Tatort:[^-(]*$" -download -path "downloads/" -download -workers 3
    
search for "Tatort"  
`-search "Tatort"`  
set the minimum duration to 20 minutes (20\*60 seconds)  
`-minduration 1200`  
limit the results to titles that match the regex `^Tatort:[^-(]*$`  
`-regex "^Tatort:[^-(]*$"`  
enable downloading  
`-download`  
save downloads to the directory  "downloads"  
`-path "downloads/"`  
allow a maximum of 3 dowloads at once  
`-workers 3`  


