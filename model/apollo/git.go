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
	// webhook 能力（service-git-webhook）：三个上游路径模板，占位符
	// {host}/{owner}/{repo}/{page}/{page_size}/{hook_id}/{delivery_id}，由
	// components/git.BuildURL 填充。空模板 = 该 provider 不支持对应能力，调用方
	// 直接返回明确错误，不静默降级（与四浏览模板同口径）。hooksPath 两家通用；
	// deliveriesPath/redeliverPath 仅 github.com（gitee v5 无 deliveries 查询与
	// 单条重发 API），sdk 侧对这两能力额外做 provider 白名单，防误配模板打错平台。
	HooksPath      string `json:"hooksPath"`
	DeliveriesPath string `json:"deliveriesPath"`
	RedeliverPath  string `json:"redeliverPath"`
}
