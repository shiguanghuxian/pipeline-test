import axios from 'axios'
import iView from 'iview'
import { Message } from 'iview'
import { Config } from "@/config"

// 请求前缀
axios.defaults.baseURL = Config.BaseUrl

// 请求之前拦截
axios.interceptors.request.use(
    config => {
        iView.LoadingBar.start();
        return config
    },
    err => {
        return Promise.reject(err)
    })


// 请求相应拦截器
axios.interceptors.response.use(
    response => {
        if (response.status == 400) {
            Message.error('req err')
        }
        iView.LoadingBar.finish();
        return response
    },
    error => {
        iView.LoadingBar.error();
        if (error.response) {
            if (error.response.status == 400) {
                Message.error(error.response.data.msg);
            }
        } else {
            Message.error('请求错误' + JSON.stringify(error));
        }
        return Promise.reject(error)
    });

export default axios