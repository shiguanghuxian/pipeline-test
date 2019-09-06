// 开发环境配置
const DEV = {
    BaseUrl: 'http://127.0.0.1:11260',
    WsBaseUrl: 'ws://127.0.0.1:11260',
}

// 正式环境
const PRO = {
    BaseUrl: '',
    WsBaseUrl: 'ws://127.0.0.1:11260',
}

export const Config = process.env.NODE_ENV === 'development' ? DEV : PRO
