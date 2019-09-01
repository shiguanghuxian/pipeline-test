function expect() {
    var code = response.getStatusCode()
    if (code == 200) {
        return true
    }
    return false
}
expect()
