# 中间件的实现

1. 首先需要c.Next()方法，用于启动调用链、传递上下文、等待剩下的中间件完成后再进行一些处理；因为 HandlerFunc
 的参数是c，所以在context实现

2. 然后是实现添加中间件的方法 RouterGroup.Use()，因为中间件是作用在分组上的，所以需要 RouterGroup 实现，但是一般是直接在engine中调用

3. 那么什么时候把合适的 middlewares 添加到 context 中呢？

应该在 engine 的 serveHTTP 中，因为 router 的 handle 方法是用于执行匹配到的路由，并且 router 在更底层访问不到上层的 RouterGroup。

engine 作为最上层的框架，serveHTTP **负责判断该请求在哪个分组中，适用于哪些 middleware** （遍历路由分组，判断前缀是否在当前请求的url中），并创建新的 context ，将这些 middleware 添加到 context 中。然后调用 router.handle() 处理这个context

最后就是 router.handle() ，也要调整，之前是直接通过路由树匹配路由，然后执行。现在为了逐层调用中间件，不能直接执行，而是先匹配路由，然后把匹配到的 handler 加入 context，调用 c.Next() 开始逐层开始执行。
注意匹配到的 handler 是最后添加的，所以会最后执行，先执行的是添加的中间件。