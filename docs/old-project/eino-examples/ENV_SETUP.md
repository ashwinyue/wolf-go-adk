# 环境变量管理

本项目使用 direnv 自动管理环境变量，并提供了 Makefile 命令来简化配置过程。

## 快速开始

如果您想快速设置环境变量管理，只需运行：

```bash
make direnv-setup
source ~/.zshrc  # 或 source ~/.bashrc
make da
```

## 手动安装 direnv

如果您想手动安装 direnv，可以使用 Homebrew 安装：

```bash
brew install direnv
```

## 配置 direnv

1. 将 direnv 钩子添加到您的 shell 配置文件中：

   对于 zsh（macOS 默认）：
   ```bash
   echo 'eval "$(direnv hook zsh)"' >> ~/.zshrc
   ```

   对于 bash：
   ```bash
   echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
   ```

2. 重新加载您的 shell 配置：
   ```bash
   source ~/.zshrc  # 或 source ~/.bashrc
   ```

## Makefile 命令

本项目提供了以下 Makefile 命令来简化环境变量管理：

- `make help` - 显示所有可用命令
- `make env` - 手动加载环境变量（不推荐，推荐使用 direnv）
- `make direnv-setup` - 自动安装和配置 direnv
- `make da` - 允许 direnv 加载 .envrc 文件

## 项目配置

项目根目录包含 `.envrc` 文件，它会自动加载 `.env` 文件中的环境变量。

首次进入项目目录时，您需要允许 direnv：

```bash
direnv allow
# 或者使用 Makefile 命令
make da
```

## 使用方法

现在，当您进入项目目录时，direnv 会自动加载环境变量。您可以直接运行任何需要这些环境变量的命令，例如：

```bash
cd adk/helloworld
go run helloworld.go
```

无需手动加载环境变量！

## 注意事项

- `.envrc` 文件应该提交到版本控制系统
- `.env` 文件包含敏感信息，不应该提交到版本控制系统
- 如果您修改了 `.envrc` 文件，direnv 会提示您重新运行 `direnv allow` 或 `make da`