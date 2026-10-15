package apollo

type GitConfig struct {
	Providers map[string]GitItemConfig `json:"providers"`
}

type GitItemConfig struct {
	Host    string `json:"host"`
	Token   string `json:"token"`
	Enable  bool   `json:"enable"`
	ShaPath string `json:"shaPath"`
	// git 页三段式只读浏览（service-git-page）：四个上游路径模板，占位符
	// {host}/{owner}/{repo}/{ref}/{path}/{commit}/{page}/{page_size}，
	// 由 components/git.BuildURL 填充。空模板 = 该 provider 不支持对应能力，
	// 调用方直接返回明确错误，不静默降级。
	FilePath     string `json:"filePath"`
	TreePath     string `json:"treePath"`
	HistoryPath  string `json:"historyPath"`
	BranchesPath string `json:"branchesPath"`
}
