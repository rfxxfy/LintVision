package stats

type FileStats struct {
	Path          string `json:"path,omitempty"`
	Ext           string `json:"ext"`
	Category      string `json:"category"`
	LinesTotal    int    `json:"lines_total"`
	LinesCode     int    `json:"lines_code"`
	LinesComments int    `json:"lines_comments"`
	LinesBlank    int    `json:"lines_blank"`
}

type ProjectStats struct {
	Files          []FileStats    `json:"files"`
	CategoryCounts map[string]int `json:"category_counts"`

	HiddenFiles   int `json:"hidden_files"`
	HiddenDirs    int `json:"hidden_dirs"`
	NonHiddenDirs int `json:"non_hidden_dirs"`
}
