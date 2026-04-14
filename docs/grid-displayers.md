# Grid Displayers 使用指南

GoAdmin 提供了 25 种内置的 Grid Displayers，用于在列表视图中以不同方式展示数据。

## 基础用法

所有 Displayers 都通过 `Column.Displayer()` 方法使用：

```go
import "github.com/zhenyangze/goadmin/grid"

grid.Column("status", "Status").
    Displayer(grid.Badge().Style(grid.BadgeStyleSuccess))
```

## Displayers 分类

### 1. 视觉展示类

#### Badge - 徽章标签
```go
grid.Column("status", "Status").
    Displayer(grid.Badge().
        Style(grid.BadgeStyleSuccess).
        Color(func(record any, value any) grid.BadgeStyle {
            if value == "active" {
                return grid.BadgeStyleSuccess
            }
            return grid.BadgeStyleWarning
        }))
```

支持样式: `default`, `primary`, `success`, `warning`, `danger`, `info`

#### Label - 标签样式
```go
grid.Column("category", "Category").
    Displayer(grid.Label().Style(grid.LabelStylePrimary))
```

#### Image - 图片预览
```go
grid.Column("avatar", "Avatar").
    Displayer(grid.Image().
        Size(50, 50).
        Preview(true))
```

#### ProgressBar - 进度条
```go
grid.Column("completion", "Completion").
    Displayer(grid.ProgressBar().
        Range(0, 100).
        ShowText(true).
        Color(func(record any, value any) string {
            if v, ok := value.(float64); ok {
                if v < 30 {
                    return "red"
                } else if v < 70 {
                    return "yellow"
                }
            }
            return "green"
        }))
```

#### QRCode - 二维码
```go
grid.Column("code", "QR Code").
    Displayer(grid.QRCode().Size(128))
```

### 2. 状态展示类

#### SwitchDisplay - 开关状态
```go
grid.Column("is_active", "Active").
    Displayer(grid.SwitchDisplay().
        Text("Yes", "No").
        Color("green", "gray"))
```

#### Checkbox - 复选框状态
```go
grid.Column("enabled", "Enabled").
    Displayer(grid.Checkbox().
        Checked(func(record any, value any) bool {
            return value == true || value == 1
        }).
        Disabled(true))
```

#### Radio - 单选状态
```go
grid.Column("priority", "Priority").
    Displayer(grid.Radio().
        OptionsMap(map[string]string{
            "low":    "Low",
            "medium": "Medium",
            "high":   "High",
        }).
        Selected(func(record any, value any) string {
            return fmt.Sprintf("%v", value)
        }))
```

#### Select - 下拉选择显示
```go
grid.Column("status", "Status").
    Displayer(grid.SelectDisplay().
        OptionsMap(map[string]string{
            "pending":  "Pending",
            "approved": "Approved",
            "rejected": "Rejected",
        }).
        Color("pending", "yellow").
        Color("approved", "green").
        Color("rejected", "red").
        DefaultColor("gray"))
```

#### SwitchGroup - 开关组
```go
grid.Column("permissions", "Permissions").
    Displayer(grid.SwitchGroup().
        Options(
            grid.Option{Value: "read", Label: "Read"},
            grid.Option{Value: "write", Label: "Write"},
            grid.Option{Value: "delete", Label: "Delete"},
        ).
        Separator(","))
```

### 3. 输入展示类

#### Input - 输入框展示
```go
grid.Column("name", "Name").
    Displayer(grid.Input().
        Type("text").
        Readonly(true).
        MaxLength(50))
```

#### Textarea - 多行文本
```go
grid.Column("description", "Description").
    Displayer(grid.Textarea().
        Rows(3).
        Cols(40).
        Readonly(true))
```

#### Editable - 行内编辑
```go
grid.Column("title", "Title").
    Displayer(grid.Editable("title").
        Type("text").
        SaveURL(func(record any) string {
            return "/admin/articles/" + getID(record) + "/quick-update"
        }))
```

### 4. 导航操作类

#### Link - 链接跳转
```go
grid.Column("website", "Website").
    Displayer(grid.Link().
        Target("_blank").
        MaxLength(30).
        URL(func(record any, value any) string {
            return fmt.Sprintf("%v", value)
        }))
```

#### Downloadable - 可下载链接
```go
grid.Column("attachment", "Attachment").
    Displayer(grid.Downloadable().
        Text("Download").
        ShowIcon(true).
        Filename(func(record any, value any) string {
            return "file.pdf"
        }))
```

#### Button - 按钮展示
```go
grid.Column("action", "Action").
    Displayer(grid.Button("View").
        Style(grid.ActionPrimary).
        URL(func(record any, value any) string {
            return "/admin/items/" + getID(record)
        }))
```

#### DropdownActions - 下拉操作菜单
```go
grid.Column("actions", "Actions").
    Displayer(grid.DropdownActions("Actions").
        Action("Edit", func(record any, value any) string {
            return "/admin/items/" + getID(record) + "/edit"
        }).
        Action("View", func(record any, value any) string {
            return "/admin/items/" + getID(record)
        }).
        ActionWithConfirm("Delete", func(record any, value any) string {
            return "/admin/items/" + getID(record) + "/delete"
        }, "Are you sure?"))
```

### 5. 布局展示类

#### Expand - 展开详情
```go
grid.Column("details", "Details").
    Displayer(grid.Expand().
        Summary("Click to expand").
        Expanded(false).
        Content(func(record any, value any) template.HTML {
            return template.HTML(fmt.Sprintf("<pre>%v</pre>", value))
        }))
```

#### Modal - 弹窗内容
```go
grid.Column("preview", "Preview").
    Displayer(grid.Modal("View").
        Title("Content Preview").
        Size("lg").
        TriggerStyle(grid.ActionPrimary).
        Content(func(record any, value any) template.HTML {
            return template.HTML(fmt.Sprintf("<div>%v</div>", value))
        }))
```

#### Limit - 文本截断
```go
grid.Column("content", "Content").
    Displayer(grid.Limit(50).
        Tooltip(true).
        Replace("..."))
```

#### Copyable - 可复制文本
```go
grid.Column("api_key", "API Key").
    Displayer(grid.Copyable().
        MaxLength(30).
        ShowIcon(true))
```

### 6. 数据展示类

#### Table - 嵌套表格
```go
grid.Column("items", "Items").
    Displayer(grid.Table().
        Columns("name", "price", "quantity").
        MaxRows(5).
        Data(func(record any) []map[string]any {
            // Return nested data
            return getItems(record)
        }))
```

#### Tree - 树形展示
```go
grid.Column("category", "Category").
    Displayer(grid.Tree().
        Label(func(item map[string]any) string {
            if name, ok := item["name"].(string); ok {
                return name
            }
            return ""
        }).
        ChildrenKey("children").
        MaxDepth(3))
```

#### DialogTree - 树形弹窗
```go
grid.Column("categories", "Categories").
    Displayer(grid.DialogTree("View Tree").
        Title("Category Hierarchy").
        Size("lg").
        TreeData(func(record any, value any) []map[string]any {
            return getCategoryTree(record)
        }))
```

#### Orderable - 排序标识
```go
grid.Column("sort_order", "Order").
    Displayer(grid.Orderable().
        Order(func(record any, value any) int {
            if v, ok := value.(int); ok {
                return v
            }
            return 0
        }).
        ShowArrows(true))
```

## 完整示例

```go
func (r *ArticleResource) Grid(grid *grid.Builder) {
    grid.
        Column("id", "ID").
            Displayer(grid.Label().Style(grid.LabelStyleDefault)).
            Sortable(true).
            Width("80px").
        Column("title", "Title").
            Displayer(grid.Link().
                URL(func(record any, value any) string {
                    return "/admin/articles/" + getID(record)
                })).
            Searchable(true).
        Column("status", "Status").
            Displayer(grid.Badge().
                Color(func(record any, value any) grid.BadgeStyle {
                    switch value {
                    case "published":
                        return grid.BadgeStyleSuccess
                    case "draft":
                        return grid.BadgeStyleWarning
                    case "archived":
                        return grid.BadgeStyleDefault
                    }
                    return grid.BadgeStyleDefault
                })).
            Filterable(true)
        Column("progress", "Progress").
            Displayer(grid.ProgressBar().Range(0, 100))
        Column("is_featured", "Featured").
            Displayer(grid.SwitchDisplay().Text("Yes", "No"))
        Column("cover_image", "Cover").
            Displayer(grid.Image().Size(60, 60).Preview(true))
        Column("tags", "Tags").
            Displayer(grid.Label().
                Color(func(record any, value any) grid.LabelStyle {
                    return grid.LabelStyleInfo
                }))
        Column("actions", "Actions").
            Displayer(grid.DropdownActions("Manage").
                Action("Edit", func(r, v any) string {
                    return "/admin/articles/" + getID(r) + "/edit"
                }).
                Action("View", func(r, v any) string {
                    return "/admin/articles/" + getID(r)
                }))
}
```

## 自定义 Displayer

你可以创建自定义的 Displayer：

```go
type MyDisplayer struct {
    prefix string
}

func (d *MyDisplayer) Display(record any, value any) template.HTML {
    return template.HTML(fmt.Sprintf("%s: %v", d.prefix, value))
}

// 使用
grid.Column("field", "Field").
    Displayer(&MyDisplayer{prefix: "Value"})
```

或使用函数式 Displayer：

```go
grid.Column("field", "Field").
    Displayer(grid.DisplayerFunc(func(record any, value any) template.HTML {
        return template.HTML(fmt.Sprintf("<b>%v</b>", value))
    }))
```