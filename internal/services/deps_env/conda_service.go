package deps_env

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"

	"baihu/internal/logger"
)

// CondaManager Conda 运行时管理器
type CondaManager struct {
	condaPath string
}

// NewCondaManager 创建 Conda 管理器
func NewCondaManager() *CondaManager {
	return &CondaManager{}
}

// GetType 获取运行时类型
func (cm *CondaManager) GetType() string {
	return "conda"
}

// IsAvailable 检查 Conda 是否可用
func (cm *CondaManager) IsAvailable() bool {
	path, err := cm.findCondaPath()
	if err != nil {
		return false
	}
	cm.condaPath = path
	return true
}

// findCondaPath 查找 conda 可执行文件路径
func (cm *CondaManager) findCondaPath() (string, error) {
	// 尝试常见的 conda 路径
	paths := []string{"conda", "micromamba", "/opt/conda/bin/conda", "/root/miniconda3/bin/conda", "/root/anaconda3/bin/conda"}
	for _, p := range paths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}
	return "", exec.ErrNotFound
}

// getCondaPath 获取 conda 路径
func (cm *CondaManager) getCondaPath() string {
	if cm.condaPath == "" {
		cm.findCondaPath()
	}
	return cm.condaPath
}

// condaEnvJSON conda env list --json 的输出结构
type condaEnvJSON struct {
	Envs []string `json:"envs"`
}

// ListEnvs 列出所有 Conda 环境
func (cm *CondaManager) ListEnvs() ([]RuntimeEnv, error) {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return nil, exec.ErrNotFound
	}

	cmd := exec.Command(condaPath, "env", "list", "--json")
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("Failed to list conda envs: %v", err)
		return nil, err
	}

	var envJSON condaEnvJSON
	if err := json.Unmarshal(output, &envJSON); err != nil {
		return nil, err
	}

	var envs []RuntimeEnv
	for _, envPath := range envJSON.Envs {
		name := extractEnvName(envPath)
		// 过滤以 . 开头的环境
		if strings.HasPrefix(name, ".") {
			continue
		}
		envs = append(envs, RuntimeEnv{
			Name:   name,
			Path:   envPath,
			Active: false,
		})
	}

	return envs, nil
}

// extractEnvName 从路径中提取环境名称
func extractEnvName(envPath string) string {
	parts := strings.Split(envPath, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		// 如果是 base 环境，路径可能是 /opt/conda 这样的
		if name == "conda" || name == "miniconda3" || name == "anaconda3" {
			return "base"
		}
		return name
	}
	return envPath
}

// CreateEnv 创建 Conda 环境
func (cm *CondaManager) CreateEnv(name string, version string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	args := []string{"create", "-n", name, "-y"}
	if version != "" {
		args = append(args, "python="+version)
	}

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to create conda env: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

// DeleteEnv 删除 Conda 环境
func (cm *CondaManager) DeleteEnv(name string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	if name == "base" {
		return nil // 不允许删除 base 环境
	}

	cmd := exec.Command(condaPath, "env", "remove", "-n", name, "-y")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to delete conda env: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

// ListPackages 列出环境中的包
func (cm *CondaManager) ListPackages(envName string) ([]RuntimePackage, error) {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return nil, exec.ErrNotFound
	}

	args := []string{"list"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.Output()
	if err != nil {
		logger.Errorf("Failed to list packages: %v", err)
		return nil, err
	}

	return parseCondaList(string(output)), nil
}

// parseCondaList 解析 conda list 输出
func parseCondaList(output string) []RuntimePackage {
	var packages []RuntimePackage
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		// 跳过注释和空行
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			name := fields[0]

			pkg := RuntimePackage{
				Name:    name,
				Version: fields[1],
			}
			if len(fields) >= 4 {
				pkg.Channel = fields[3]
			}
			packages = append(packages, pkg)
		}
	}

	return packages
}

// InstallPackage 安装包
func (cm *CondaManager) InstallPackage(envName string, packageName string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	args := []string{"install", "-y"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}
	args = append(args, packageName)

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to install package: %v, output: %s", err, string(output))
		return err
	}

	return nil
}

// UninstallPackage 卸载包
func (cm *CondaManager) UninstallPackage(envName string, packageName string) error {
	condaPath := cm.getCondaPath()
	if condaPath == "" {
		return exec.ErrNotFound
	}

	args := []string{"remove", "-y"}
	if envName != "" && envName != "base" {
		args = append(args, "-n", envName)
	}
	args = append(args, packageName)

	cmd := exec.Command(condaPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("Failed to uninstall package: %v, output: %s", err, string(output))
		return err
	}

	return nil
}
