当前项目参考dcat-admin ,使用golang 实现类似的框架,golang 主要是gorm+tailwindcss(和参考框架架构不一
  样) ,请仔细思考规划,并通过demo(比如gin+sqllite)的方式测试最终效果,希望能完美复刻整个模块,制作成go 包,可以供其他框架使用

当前优先级：先把基于现有框架的完整可运行 demo 做扎实，保证开箱即跑、可验证、可演示。

后续待办（已备注，当前不优先）：
1. relation-backed nested editor 继续泛化（更多子集合/更复杂映射）
2. relation / field / repository hook 扩展点继续增强
3. upload 云存储、图片处理、缩略图等高级能力
