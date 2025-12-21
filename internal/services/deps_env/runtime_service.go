package deps_env

// RuntimeEnv 运行时环境信息
type RuntimeEnv struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

// RuntimePackage 包信息
type RuntimePackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Channel string `json:"channel,omitempty"`
}

// RuntimeManager 运行时管理器接口
type RuntimeManager interface {
	// GetType 获取运行时类型
	GetType() string
	// IsAvailable 检查运行时是否可用
	IsAvailable() bool
	// ListEnvs 列出所有环境
	ListEnvs() ([]RuntimeEnv, error)
	// CreateEnv 创建环境
	CreateEnv(name string, version string) error
	// DeleteEnv 删除环境
	DeleteEnv(name string) error
	// ListPackages 列出环境中的包
	ListPackages(envName string) ([]RuntimePackage, error)
	// InstallPackage 安装包
	InstallPackage(envName string, packageName string) error
	// UninstallPackage 卸载包
	UninstallPackage(envName string, packageName string) error
}

// RuntimeService 运行时服务
type RuntimeService struct {
	managers map[string]RuntimeManager
}

// NewRuntimeService 创建运行时服务
func NewRuntimeService() *RuntimeService {
	rs := &RuntimeService{
		managers: make(map[string]RuntimeManager),
	}
	// 注册 Conda 管理器
	rs.RegisterManager(NewCondaManager())
	return rs
}

// RegisterManager 注册运行时管理器
func (rs *RuntimeService) RegisterManager(manager RuntimeManager) {
	rs.managers[manager.GetType()] = manager
}

// GetManager 获取指定类型的管理器
func (rs *RuntimeService) GetManager(runtimeType string) RuntimeManager {
	return rs.managers[runtimeType]
}

// GetAvailableRuntimes 获取可用的运行时列表
func (rs *RuntimeService) GetAvailableRuntimes() []string {
	var available []string
	for name, manager := range rs.managers {
		if manager.IsAvailable() {
			available = append(available, name)
		}
	}
	return available
}
