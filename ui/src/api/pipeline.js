import axios from './api'

const Pipeline = {
    /**
     * 全部角色
     */
    GetList(userId, username, roleId, page, pageSize) {
        // return axios.get(`/v1/user?user_id=${userId}&name=${username}&role_id=${roleId}&page=${page}&page_size=${pageSize}`);
    },

    /**
     * 添加
     * @param {*} data 添加角色信息
     */
    Add(data) {
        return axios.post('/v1/user', data)
    },

    /**
     * 运行一个流水线
     * @param {*} pipelineId 
     * @param {*} projectId 
     */
    RunTask(pipelineId, projectId) {
        if(!pipelineId || !projectId){
            console.log('参数错误');
            
            return
        }
        return axios.get(`/v1/pipeline/runTask?pipeline_id=${pipelineId}&project_id=${projectId}`)
    }



}

export {
    Pipeline
}
