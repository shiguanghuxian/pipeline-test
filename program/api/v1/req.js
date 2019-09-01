function setHeader() {
    // 获取csrf cookie
    var csrfSecret = request.getCookie('csrfSecret')
    var ok = request.header('X-CSRF-TOKEN', csrfSecret)
	return ok
}
setHeader()
