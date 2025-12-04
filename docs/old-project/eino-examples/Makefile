# Eino Examples Makefile

.PHONY: help env direnv-setup da

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  env           - 加载环境变量"
	@echo "  direnv-setup  - 安装和配置direnv"
	@echo "  da            - 允许direnv加载.envrc文件"

# 加载环境变量
env:
	@echo "加载环境变量..."
	@export $$(grep -v '^#' .env | xargs) && echo "环境变量已加载"

# 安装和配置direnv
direnv-setup:
	@echo "安装direnv..."
	@brew install direnv || echo "direnv可能已安装或使用其他包管理器安装"
	@echo "配置direnv钩子..."
	@if [ -n "$$ZSH_VERSION" ]; then \
		grep -q 'eval "$$(direnv hook zsh)"' ~/.zshrc || echo 'eval "$$(direnv hook zsh)"' >> ~/.zshrc; \
	elif [ -n "$$BASH_VERSION" ]; then \
		grep -q 'eval "$$(direnv hook bash)"' ~/.bashrc || echo 'eval "$$(direnv hook bash)"' >> ~/.bashrc; \
	fi
	@echo "direnv安装和配置完成!"
	@echo "请运行 'source ~/.zshrc' 或 'source ~/.bashrc' 重新加载shell配置"

# 允许direnv加载.envrc文件
da:
	@echo "允许direnv加载.envrc文件..."
	@direnv allow
	@echo "direnv已配置完成，现在进入项目目录时会自动加载环境变量"