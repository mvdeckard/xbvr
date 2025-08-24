package scrape

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mozillazg/go-slugify"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/models"
)

type wetVRRelease struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CachedSlug  string   `json:"cachedSlug"`
	ReleasedAt  string   `json:"releasedAt"`
	PosterUrl   string   `json:"posterUrl"`
	ThumbUrls   []string `json:"thumbUrls"`
	TrailerUrl  string   `json:"trailerUrl"`
	Actors      []struct {
		Name string `json:"name"`
	} `json:"actors"`
	DownloadOptions []struct {
		Quality  string `json:"quality"`
		Filename string `json:"filename"`
	} `json:"downloadOptions"`
}

// Scene:Duration map
var wetVRDurations = map[int]int{
    67991: 2885, 67964: 2899, 67975: 2142, 67993: 2130, 67972: 2650, 67989: 1862, 67965: 2920, 67971: 1872,
    68144: 2280, 68117: 2000, 68100: 2171, 68633: 2573, 68090: 2750, 68086: 2187, 68077: 1953, 68020: 2125,
    67974: 2879, 68011: 2926, 67973: 2301, 67976: 2949, 68304: 2872, 68294: 2344, 68277: 2580, 68265: 2340,
    68248: 2833, 68233: 2575, 68216: 2913, 68203: 2580, 68190: 2535, 68180: 2747, 68158: 2514, 68132: 2682,
    68613: 2865, 68591: 2999, 68569: 3616, 68537: 2890, 68513: 2105, 68434: 2595, 68402: 2524, 68377: 2508,
    68360: 1946, 68354: 2555, 68332: 2647, 68321: 2899, 68994: 3196, 68973: 2644, 68931: 2788, 68888: 2740,
    68778: 2658, 68742: 2172, 68762: 3004, 68755: 2719, 68724: 2636, 68693: 3013, 68666: 3228, 68651: 2714,
    69220: 2844, 69199: 2429, 69177: 2314, 69159: 1798, 69141: 2997, 69115: 2555, 69108: 3112, 69077: 2212,
    69055: 2910, 69037: 2143, 69013: 2359, 69514: 2216, 69491: 2182, 69476: 3034, 69446: 2714, 69453: 2918,
    69405: 1746, 69385: 2169, 69356: 2176, 69337: 1909, 69313: 3475, 69297: 1805, 69274: 2041, 69778: 3165,
    69749: 3118, 69728: 2511, 69701: 2499, 69670: 3557, 69648: 1741, 69625: 2324, 69606: 2346, 69586: 2035,
    69565: 2220, 69539: 2520, 73706: 2247, 73669: 3261, 73612: 2506, 73562: 3246, 73519: 3652, 73458: 2236,
    73409: 3045, 73366: 2434, 73323: 3204, 73262: 2751, 73213: 1923, 73167: 3431, 74217: 3359, 74179: 2612,
    74136: 3321, 74092: 2111, 74049: 2379, 74010: 2900, 73971: 2536, 73921: 2578, 73878: 2310, 73837: 2265,
    73791: 2852, 73762: 2854, 74665: 2319, 74629: 2199, 74596: 2295, 74555: 2526, 74519: 1983, 74478: 2353,
    74442: 2062, 74416: 3178, 74370: 2980, 74338: 2467, 74297: 2851, 74257: 3218, 75091: 2675, 75059: 2164,
    75024: 2541, 74990: 2833, 74962: 2776, 74928: 2825, 74856: 3370, 74820: 2306, 74776: 2910, 74741: 3026,
    74710: 2418, 75161: 2599, 75128: 2616, 69241: 2385, 75195: 2701, 75232: 2827, 75486: 2795, 75445: 2467,
    75427: 2174, 75368: 2091, 75333: 2429, 75289: 2147, 75299: 3016, 75515: 2623, 75540: 2576, 75568: 3292,
    75594: 2627, 75625: 1871, 74889: 2344, 75646: 2168, 75677: 2630, 75734: 2832, 75701: 3131, 75764: 2642,
    75792: 1812, 75808: 2836, 75846: 2152,
}

type wetVRReleaseList struct {
	Items      []wetVRRelease `json:"items"`
	Pagination struct {
		NextPage   string `json:"nextPage"`
		TotalItems int    `json:"totalItems"`
		TotalPages int    `json:"totalPages"`
	} `json:"pagination"`
}

func fetchJSON(url string, target interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	httpConfig := GetCoreDomain(url) + "-scraper"
	log.Debugf("Using Header/Cookies from %s", httpConfig)
	SetupHtmlRequest(httpConfig, req)

	req.Header.Set("x-site", "wetvr.com")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the JSON
	return json.Unmarshal(body, target)
}

func WetVR(wg *models.ScrapeWG, updateSite bool, knownScenes []string, out chan<- models.ScrapedScene, singleSceneURL string, singeScrapeAdditionalInfo string, limitScraping bool) error {
	defer wg.Done()
	scraperID := "wetvr"
	siteID := "WetVR"
	logScrapeStart(scraperID, siteID)

	processScene := func(scene wetVRRelease) {
		baseURL := "https://wetvr.com"
		sceneURL := fmt.Sprintf("%s/video/%s", baseURL, scene.CachedSlug)

		// Skip if scene already exists in database
		if funk.ContainsString(knownScenes, sceneURL) {
			return
		}

		sc := models.ScrapedScene{}
		sc.ScraperID = scraperID
		sc.SceneType = "VR"
		sc.Studio = "WetVR"
		sc.Site = siteID
		sc.Title = scene.Title
		sc.HomepageURL = sceneURL
		sc.MembersUrl = strings.Replace(sceneURL, baseURL+"/", baseURL+"/members/", 1)
		sc.SiteID = fmt.Sprintf("%d", scene.ID)
		sc.SceneID = slugify.Slugify(sc.Site + "-" + sc.SiteID)

		// Set duration from our mapping (seconds)
		if duration, ok := wetVRDurations[scene.ID]; ok {
			sc.Duration = duration
		}

		// Convert release date format
		if t, err := time.Parse(time.RFC3339, scene.ReleasedAt); err == nil {
			sc.Released = t.Format("2006-01-02")
		}
		sc.Synopsis = scene.Description

		// Cover image
		if scene.PosterUrl != "" {
			sc.Covers = append(sc.Covers, scene.PosterUrl)
		}

		// Gallery
		sc.Gallery = scene.ThumbUrls

		// Cast
		for _, actor := range scene.Actors {
			sc.Cast = append(sc.Cast, actor.Name)
		}

		// Trailer
		if scene.TrailerUrl != "" {
			sc.TrailerType = "url"
			sc.TrailerSrc = scene.TrailerUrl
		}

		// Filenames from downloadOptions
		for _, opt := range scene.DownloadOptions {
			sc.Filenames = append(sc.Filenames, opt.Filename)
		}

		out <- sc
	}

	if singleSceneURL != "" {
		// Extract slug from URL - handle both old and new URL formats
		slug := singleSceneURL
		if strings.Contains(singleSceneURL, "/video/") {
			slug = strings.TrimPrefix(singleSceneURL, "https://wetvr.com/video/")
		}
		// Remove any trailing slashes and query parameters
		slug = strings.Split(slug, "?")[0]
		slug = strings.TrimSuffix(slug, "/")
		// Get the last part of the path as the slug
		parts := strings.Split(slug, "/")
		slug = parts[len(parts)-1]

		apiURL := fmt.Sprintf("https://wetvr.com/api/releases/%s", slug)

		var scene wetVRRelease
		if err := fetchJSON(apiURL, &scene); err != nil {
			log.Error(err)
			return err
		}

		if scene.ID == 0 {
			log.Errorf("[%s] Failed to get valid scene data for %s", scraperID, apiURL)
			return fmt.Errorf("invalid scene data received")
		}

		processScene(scene)
	} else {
		page := 1
		sceneCount := 0

		for {
			apiURL := fmt.Sprintf("https://wetvr.com/api/releases?sort=latest&page=%d", page)
			// Skip per-page logging

			var releases wetVRReleaseList
			if err := fetchJSON(apiURL, &releases); err != nil {
				log.Error(err)
				return err
			}

			// Skip per-page scene count logging
			if len(releases.Items) == 0 {
				break
			}

			for _, scene := range releases.Items {
				// Hard rule: skip scene 75122 (duplicate of 75091)
				if scene.ID == 75122 {
					continue
				}
				if scene.ID != 0 {
					processScene(scene)
					sceneCount++
				}
			}

			if limitScraping {
				break
			}
			page++
		}

		log.Infof("[%s] Successfully scraped %d new scenes", scraperID, sceneCount)
	}

	if updateSite {
		updateSiteLastUpdate(scraperID)
	}
	logScrapeFinished(scraperID, siteID)
	return nil
}

func init() {
	registerScraper("wetvr", "WetVR", "https://wetvr.com/images/sites/wetvr/wetvr-favicon.ico", "wetvr.com", WetVR)
}
