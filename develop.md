# Gogit 开发学习路线：持久 Shell + PTY

## 1. 最终目标

Gogit 不是重新实现 PowerShell、Bash 或 Git，而是在真实 Shell 外增加一层交互能力：

```text
键盘输入
   ↓
Gogit 行编辑器
   ├── 维护输入内容和光标
   ├── 分析 Git 命令
   └── 显示候选项及解释
   ↓ Enter
持久 PTY 会话
   ↓
PowerShell / Bash / Zsh
   ↓
Git 和其他系统程序
```

第一阶段的目标不是立即实现补全菜单，而是先证明下面这些状态能够在多条命令之间保留：

- 当前工作目录；
- Shell 环境变量；
- PowerShell 函数或 Bash 函数；
- alias；
- Shell 自身的其他会话状态。

## 2. 为什么需要 PTY

当前实现对每条命令执行一次 `shell -c command`，因此每条命令都在新的 Shell 进程中运行。子进程结束后，它的目录、变量和函数也随之消失。

持久 Shell 表示 Gogit 启动时只创建一个 Shell，后续所有命令都交给同一个进程。PTY（Pseudo Terminal，伪终端）则让这个 Shell 以及它启动的程序认为自己连接着真正的终端，从而保留颜色、光标控制和交互式程序行为。

不同平台使用不同的系统机制：

- Windows：ConPTY；
- Linux/macOS：Unix PTY；
- Gogit 上层只依赖统一的会话接口，不直接依赖平台细节。

## 3. 当前技术基线

项目的 `go.mod` 当前面向 Go 1.25.0，并直接依赖 `github.com/charmbracelet/x/xpty` v0.1.4。`cmd/root.go` 已接入持久 PTY Shell：上层通过 `session.ShellSession` 工作，内部实现为基于 xpty 的 `ptySession`；Windows 使用 ConPTY，Linux/macOS 使用 Unix PTY。

需要注意：

- 该项目明确将 `x` 仓库中的包定义为实验性包，不保证向后兼容；
- 当前版本要求 Go 1.25；
- 集成后需要持续验证 Shell 生命周期、错误传播、终端恢复和窗口 resize；
- Git 自动补全仍是后续计划功能。

备选库是 `github.com/aymanbagabas/go-pty`，它同样支持 Unix PTY 和 Windows ConPTY，但当前版本也要求 Go 1.25。

`github.com/creack/pty` 只适合 Unix，不能单独承担 Gogit 的原生 Windows PTY 实现。

## 4. 建议的代码边界

随着项目增长，职责可以逐步拆分为：

```text
main.go
   └── 只负责启动 Gogit

cmd/
   └── 组织应用生命周期

internal/session/
   ├── 定义 ShellSession 抽象
   ├── 创建和关闭持久 Shell
   ├── 读写 PTY
   └── 调整 PTY 窗口大小

internal/terminal/
   ├── 读取键盘事件
   ├── 管理 raw mode
   └── 绘制输入行及候选菜单

internal/protocol/
   ├── 给命令附加完成标记
   ├── 从输出中识别完成标记
   └── 获取退出码和当前目录

internal/suggest/
   ├── 解析光标附近的 token
   ├── 生成 Git 候选项
   └── 提供候选项解释

internal/app/
   └── 协调输入、补全、会话和界面状态
```

这些目录不需要一次性全部创建。只在当前学习阶段确实需要时再增加。

## 5. ShellSession 应负责什么

会话抽象最终至少需要表达这些能力：

- 启动一个 Shell；
- 从 PTY 读取输出；
- 向 PTY 写入输入；
- 调整终端宽度和高度；
- 等待 Shell 退出；
- 关闭 PTY 并释放资源。

上层应用不应该到处判断 `runtime.GOOS`。平台差异应当被限制在会话创建层或依赖库内部。

## 6. 分阶段学习与实现

### 阶段 0：理解当前基线

学习目标：明确当前 Gogit 的父子进程关系。

需要能够解释：

- `main.go` 为什么调用 `cmd.Run()`；
- `os.Stdin`、`os.Stdout`、`os.Stderr` 分别是什么；
- `exec.Command(shell, "-c", command)` 为什么每次都会创建新 Shell；
- 为什么当前的 `cd` 必须由 Gogit 自己处理。

完成标准：不看代码也能画出“Gogit → 临时 Shell → 系统命令”的进程图。

### 阶段 1：做一个纯 PTY 透传实验

学习目标：先让一个 Shell 在 PTY 中持续运行，不加入自定义提示符和补全逻辑。

这一阶段的数据流是：

```text
os.Stdin  ──复制──> PTY ──> Shell
os.Stdout <─复制── PTY <── Shell
```

实现任务：

1. 做出明确的 Go 工具链版本选择；
2. 引入选定的跨平台 PTY 依赖；
3. 创建一个 PTY；
4. 在 PTY 中启动一次 PowerShell 或 Bash；
5. 用两个方向的数据复制连接当前终端和 PTY；
6. 等待 Shell 退出并正确关闭资源。

Windows 下优先寻找 `pwsh.exe`，不存在时再使用 `powershell.exe`。Unix 下读取 `SHELL` 环境变量，不存在时使用合理的默认 Shell。

验收命令：

```text
设置一个变量
读取这个变量
cd 到另一个目录
查看当前目录
定义一个函数或 alias
在下一条命令调用它
git status
exit
```

完成标准：以上状态在同一次 Gogit 运行期间保持，并且 `exit` 后 Gogit 能正常退出。

### 阶段 2：理解终端 raw mode

学习目标：理解为什么普通按行读取不足以实现即时补全。

普通终端通常先在终端驱动层编辑一整行，按 Enter 后程序才收到内容。raw mode 会让 Gogit 立刻收到每次按键，包括方向键、Tab、Backspace 和 Ctrl+C。

实现任务：

1. 进入 raw mode；
2. 逐字节或逐事件观察键盘输入；
3. 程序退出时恢复终端原状态；
4. 程序异常返回时也必须恢复终端；
5. 识别普通字符和常见控制键。

完成标准：按键可以被 Gogit 立即识别，退出后终端的回显和 Enter 行为仍然正常。

### 阶段 3：建立命令完成协议

学习目标：在 Shell 不退出的情况下判断“一条命令已经结束”。

Gogit 向 Shell 发送用户命令时，需要附加一个本次会话唯一的完成标记。Shell 完成命令后输出标记和退出码。Gogit 读取到完整标记后，隐藏协议内容并重新显示自己的提示符。

需要处理：

- 标记可能被拆成多个读取块；
- 一次读取中可能包含普通输出和标记；
- 用户命令可能失败；
- 标记必须包含随机会话 ID，降低与普通程序输出冲突的可能；
- PowerShell 与 POSIX Shell 的命令拼接语法不同。

完成标准：成功命令、失败命令和大量输出命令都能被正确判断结束，并得到正确退出码。

### 阶段 4：实现最小行编辑器

学习目标：让 Gogit 而不是 Shell 管理待输入的命令行。

第一版只实现：

- 插入普通字符；
- Backspace；
- 左右移动光标；
- Enter 提交；
- Ctrl+C 清空当前输入；
- Ctrl+D 在空输入时退出；
- 上下键浏览历史。

需要把“字符串字节位置”“Unicode 字符位置”和“终端显示宽度”区分开。中文字符通常不是一个字节，显示宽度也不一定等于字符数量。

完成标准：可以稳定编辑包含中文和英文的命令，提示符不会被重复输出破坏。

### 阶段 5：处理交互式子程序

学习目标：支持 `vim`、`less`、`top`、`ssh` 等会接管终端的程序。

Gogit 至少需要两种状态：

```text
编辑模式：Gogit 拦截按键，用于补全和编辑
透传模式：按键和输出直接在终端与 PTY 之间传递
```

这一阶段是项目中较困难的部分。不能只根据命令名字维护一份交互程序清单，需要研究 Shell/终端状态变化或设计明确的切换策略。

完成标准：进入和退出至少一个全屏终端程序后，Gogit 的提示符和输入状态可以恢复。

### 阶段 6：实现 Git token 解析

学习目标：根据光标位置判断用户正在输入什么。

需要逐步支持：

- `git` 一级命令；
- `git branch` 等子命令参数；
- 长参数与短参数；
- 引号中的内容；
- 光标位于命令中间；
- `|`、重定向和命令连接符前后的边界。

不要一开始尝试完整解析所有 PowerShell/Bash 语法。先限定为单条简单 Git 命令，并明确不支持的语法。

完成标准：给定输入文本和光标位置，解析器能返回当前 token、前一个 token 和所属 Git 子命令。

### 阶段 7：候选项与解释

学习目标：把解析结果转换为用户能理解的建议。

候选项数据至少包含：

- 实际插入的文本；
- 显示名称；
- 简短解释；
- 候选类型，如子命令、参数、分支或文件；
- 排序信息。

候选来源分为：

- 静态目录：Git 子命令、参数及解释；
- 动态查询：本地分支、远程分支、tag、remote；
- 文件系统：路径和文件名；
- 历史信息：最近使用的候选。

完成标准：输入 `git branch --sh` 时能展示 `--show-current`，并在附近显示其作用说明。

### 阶段 8：可靠性与跨平台收尾

需要补齐：

- 终端窗口尺寸变化时调用 PTY resize；
- Ctrl+C、Ctrl+Break 和 Unix 信号；
- Shell 意外退出；
- PTY 读取协程退出；
- 重复关闭资源；
- Windows 与 Unix 换行差异；
- ANSI 转义序列；
- 单元测试、集成测试和手工终端测试。

完成标准：正常退出、异常退出和 Ctrl+C 后终端都不会留在损坏状态。

## 7. 并发设计原则

PTY 天然涉及并发：一边等待键盘输入，一边持续读取 Shell 输出。实现时遵守以下规则：

- PTY 输出只由一个读取循环消费；
- 多个功能需要输出时，由读取循环分发事件，不要竞争读取；
- 多处写 PTY 时需要串行化，避免命令和按键字节交错；
- 用清晰的关闭信号结束协程；
- 不依赖关闭顺序碰巧正确；
- 错误需要带上“创建 PTY、启动 Shell、读取、写入或等待”等上下文。

## 8. 每一步的学习方式

每一个阶段都遵循相同循环：

1. 先画数据流和进程关系；
2. 只写能验证当前概念的最小改动；
3. 编译；
4. 手工执行验收命令；
5. 解释观察到的现象；
6. 再重构代码；
7. 提交一个小 Git commit；
8. 进入下一阶段。

不要同时实现 PTY、raw mode、命令协议和补全菜单。否则遇到输入或输出错乱时，很难定位是哪一层造成的。

## 9. 当前下一步

当前持久 PTY 基线已经完成，下一步按以下顺序推进：

1. 完成 `ptySession` 的并发生命周期与子进程回收；
2. 传播终端输入、输出和恢复错误；
3. 将宿主终端 resize 转发给 PTY；
4. 为生命周期和错误路径增加测试；
5. 再进入行编辑器和 Git 自动补全。

## 10. 参考资料

- Charmbracelet xpty 文档：https://pkg.go.dev/github.com/charmbracelet/x/xpty
- Charmbracelet x 仓库：https://github.com/charmbracelet/x
- go-pty 仓库：https://github.com/aymanbagabas/go-pty
- Unix PTY 库说明：https://github.com/creack/pty
- Go 的 Windows ConPTY 支持讨论：https://github.com/golang/go/issues/62708
- Microsoft ConPTY 文档：https://learn.microsoft.com/windows/console/creating-a-pseudoconsole-session
