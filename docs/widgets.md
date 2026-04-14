# Widgets 使用指南

GoAdmin 提供了 22 种内置 Widgets，用于构建丰富的管理界面。

## Widgets 分类

### 表单组件

#### Checkbox - 复选框
```go
import "github.com/zhenyangze/goadmin/widgets/checkbox"

w := checkbox.New("permissions").
    Options(
        checkbox.Option{Value: "read", Label: "Read"},
        checkbox.Option{Value: "write", Label: "Write"},
        checkbox.Option{Value: "delete", Label: "Delete"},
    ).
    SetChecked([]string{"read", "write"}).
    Inline(true)

html := w.RenderTo()
```

#### Radio - 单选框
```go
import "github.com/zhenyangze/goadmin/widgets/radio"

w := radio.New("status").
    OptionsMap(map[string]string{
        "active":   "Active",
        "inactive": "Inactive",
        "pending":  "Pending",
    }).
    Selected("active").
    Inline(true)

html := w.RenderTo()
```

### 数据展示

#### Table - 表格
```go
import "github.com/zhenyangze/goadmin/widgets/table"

w := table.New().
    ID("my-table").
    Column("name", "Name").
    Column("email", "Email").
    Column("role", "Role").
    Row(map[string]template.HTML{
        "name":  template.HTML("John"),
        "email": template.HTML("john@example.com"),
        "role":  template.HTML("Admin"),
    }).
    Row(map[string]template.HTML{
        "name":  template.HTML("Jane"),
        "email": template.HTML("jane@example.com"),
        "role":  template.HTML("User"),
    }).
    Striped(true).
    Hover(true)

html := w.RenderTo()
```

简单表格：
```go
headers := []string{"Name", "Age", "City"}
data := [][]string{
    {"John", "30", "New York"},
    {"Jane", "25", "London"},
}
w := table.SimpleTable(headers, data)
```

#### LazyTable - 懒加载表格
```go
import "github.com/zhenyangze/goadmin/widgets/lazytable"

w := lazytable.New("users-table").
    LoadURL("/api/users").
    Column("id", "ID").
    Column("name", "Name").
    Column("email", "Email").
    PageSize(20).
    Striped(true)

html := w.RenderTo()
```

### 内容展示

#### Terminal - 终端输出
```go
import "github.com/zhenyangze/goadmin/widgets/terminal"

// 程序化创建
w := terminal.New().
    ID("build-log").
    Theme("dark").
    Height("400px").
    AddCommand("npm install").
    AddInfo("Installing dependencies...").
    AddSuccess("Dependencies installed successfully!").
    AddWarning("Deprecated package detected").
    AddError("Build failed: missing module")

html := w.RenderTo()
```

从静态内容创建：
```go
content := `$ npm install
Installing packages...
[OK] Installed 42 packages
$ npm run build
Building application...
[ERROR] Module not found: './components/Button'`

w := terminal.Static(content)
```

主题：支持 `dark` 和 `light`

#### Card - 卡片容器
```go
import "github.com/zhenyangze/goadmin/widgets/card"

w := card.New().
    Title("User Statistics").
    Subtitle("Monthly overview").
    Content(template.HTML("<p>Content here</p>")).
    Footer(template.HTML("<a href=\"#\">View Details</a>")).
    Collapsible(true).
    Loading(false)

html := w.RenderTo()
```

#### Box - 容器组件
```go
import "github.com/zhenyangze/goadmin/widgets/box"

w := box.New().
    Title("Information").
    Content(template.HTML("<p>Box content</p>")).
    Solid(true).
    Style(box.StylePrimary)
```

### 导航组件

#### Tab - 标签页
```go
import "github.com/zhenyangze/goadmin/widgets/tab"

w := tab.New().
    Add("General", template.HTML("<p>General settings</p>")).
    Add("Security", template.HTML("<p>Security settings</p>")).
    Add("Notifications", template.HTML("<p>Notification preferences</p>")).
    Active(0)

html := w.RenderTo()
```

#### Dropdown - 下拉菜单
```go
import "github.com/zhenyangze/goadmin/widgets/dropdown"

w := dropdown.New("Actions").
    ButtonStyle(dropdown.StylePrimary).
    Item("Edit", "/edit").
    Item("Delete", "/delete").
    Divider().
    Item("Export", "/export")

html := w.RenderTo()
```

#### Tree - 树形组件
```go
import "github.com/zhenyangze/goadmin/widgets/tree"

w := tree.New().
    ID("category-tree").
    Data([]tree.Node{
        {
            ID:       "1",
            Label:    "Electronics",
            Children: []tree.Node{
                {ID: "1-1", Label: "Phones"},
                {ID: "1-2", Label: "Laptops"},
            },
        },
        {
            ID:    "2",
            Label: "Clothing",
        },
    }).
    ExpandAll(true)

html := w.RenderTo()
```

### 反馈组件

#### Alert - 提示框
```go
import "github.com/zhenyangze/goadmin/widgets/alert"

w := alert.New("Operation completed successfully!").
    Type(alert.TypeSuccess).
    Dismissible(true).
    Icon(true)

html := w.RenderTo()
```

类型: `success`, `info`, `warning`, `danger`

#### Callout - 提示框
```go
import "github.com/zhenyangze/goadmin/widgets/callout"

w := callout.New().
    Title("Important Notice").
    Content(template.HTML("<p>Please review the changes before saving.</p>")).
    Style(callout.StyleWarning)

html := w.RenderTo()
```

#### Tooltip - 工具提示
```go
import "github.com/zhenyangze/goadmin/widgets/tooltip"

w := tooltip.New("Hover me").
    Content("This is the tooltip content").
    Position(tooltip.PositionTop)

html := w.RenderTo()
```

### 工具组件

#### Code - 代码高亮
```go
import "github.com/zhenyangze/goadmin/widgets/code"

w := code.New(`func main() {
    fmt.Println("Hello, World!")
}`).
    Language("go").
    LineNumbers(true)

html := w.RenderTo()
```

#### Markdown - Markdown 渲染
```go
import "github.com/zhenyangze/goadmin/widgets/markdown"

w := markdown.New(`# Hello World

This is **bold** and *italic* text.

- Item 1
- Item 2
- Item 3
`)

html := w.RenderTo()
```

#### Dump - 数据调试
```go
import "github.com/zhenyangze/goadmin/widgets/dump"

// 调试变量
data := map[string]any{
    "name": "John",
    "age":  30,
}

w := dump.New(data)
html := w.RenderTo()
```

#### DarkModeSwitcher - 深色模式切换
```go
import "github.com/zhenyangze/goadmin/widgets/darkmode"

w := darkmode.New().
    DefaultMode(darkmode.ModeAuto)

html := w.RenderTo()
```

### 异步组件

#### Async - 异步加载
```go
import "github.com/zhenyangze/goadmin/widgets/async"

w := async.New().
    ID("stats-widget").
    LoadURL("/api/dashboard/stats").
    LoadingText("Loading statistics...")

html := w.RenderTo()
```

#### Lazy - 懒加载
```go
import "github.com/zhenyangze/goadmin/widgets/lazy"

w := lazy.New().
    ID("heavy-content").
    LoadURL("/api/heavy-data").
    Placeholder(template.HTML("<div>Loading...</div>"))

html := w.RenderTo()
```

### 图表组件

#### Chart - 基础图表
```go
import "github.com/zhenyangze/goadmin/widgets/chart"

w := chart.New().
    Type(chart.TypeLine).
    Data(chart.Data{
        Labels: []string{"Jan", "Feb", "Mar"},
        Datasets: []chart.Dataset{
            {
                Label: "Sales",
                Data:  []float64{100, 200, 150},
            },
        },
    })

html := w.RenderTo()
```

## 在 Dashboard 中使用 Widgets

```go
func (d *Dashboard) Content() template.HTML {
    var widgets []template.HTML

    // 统计卡片
    statsCard := card.New().
        Title("Total Users").
        Content(template.HTML(`
            <div class="text-3xl font-bold">1,234</div>
            <div class="text-green-500 text-sm">+12% from last month</div>
        `))
    widgets = append(widgets, statsCard.RenderTo())

    // 最近活动表格
    activityTable := table.New().
        Column("time", "Time").
        Column("user", "User").
        Column("action", "Action")

    for _, log := range d.recentLogs {
        activityTable.Row(map[string]template.HTML{
            "time":   template.HTML(log.Time),
            "user":   template.HTML(log.User),
            "action": template.HTML(log.Action),
        })
    }
    widgets = append(widgets, activityTable.RenderTo())

    // 终端输出
    term := terminal.New().
        Height("200px").
        AddInfo("System started").
        AddSuccess("Connected to database").
        AddCommand("SELECT * FROM users")
    widgets = append(widgets, term.RenderTo())

    // 合并所有 widgets
    var result strings.Builder
    for _, w := range widgets {
        result.WriteString(string(w))
    }
    return template.HTML(result.String())
}
```

## 自定义 Widget

创建自定义 Widget 只需要实现简单的接口：

```go
package mywidget

import "html/template"

type MyWidget struct {
    title string
    content template.HTML
}

func New(title string) *MyWidget {
    return &MyWidget{title: title}
}

func (w *MyWidget) Content(content template.HTML) *MyWidget {
    w.content = content
    return w
}

func (w *MyWidget) RenderTo() template.HTML {
    return template.HTML(fmt.Sprintf(`
        <div class="my-widget">
            <h3>%s</h3>
            <div class="content">%s</div>
        </div>
    `, w.title, w.content))
}
```

## 在 Form 中使用 Widgets

```go
func (r *MyResource) Form(form *form.Builder) {
    form.Field(
        form.Custom("custom_field", "Custom").
            Widget(func() template.HTML {
                return checkbox.New("options").
                    OptionsMap(map[string]string{
                        "opt1": "Option 1",
                        "opt2": "Option 2",
                    }).
                    RenderTo()
            }),
    )
}
```