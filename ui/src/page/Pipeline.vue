
<template>
  <div class="pipeline">
    <Card>
      <Table border :columns="columns" :data="data"></Table>
      <div class="page">
        <Page
          @on-change="changeListPage"
          @on-page-size-change="pageSizeChange"
          :total="total"
          :current="page"
          :page-size="pageSize"
          show-sizer
        />
      </div>
    </Card>

    <!-- 运行测试弹框 -->
    <Drawer
      title="Run task"
      v-model="showRunTask"
      width="80"
      :mask-closable="false"
      :styles="styles"
    >
      <div class="drawer-body">
        <Split v-model="run_split">
          <div slot="left" class="split-pane">
            <Timeline>
              <TimelineItem
                :color="item.is_expect == true ? 'green' : item.is_expect == undefined ? 'blue' : 'red'"
                v-for="(item, key) in taskList"
                :key="key"
              >
                <p class="line-title">{{ item.name }}</p>
                <p class="line-content">这里未来会显示url之类</p>
              </TimelineItem>
            </Timeline>
          </div>
          <div slot="right" class="split-pane" id="new_message">
            <Alert
              :type="getAleryType(item, key)"
              v-for="(item, key) in taskLogs"
              :key="key"
            >{{ item.date_str }} : <b>{{ item.msg }}</b></Alert>
          </div>
        </Split>
      </div>

      <div class="drawer-footer">
        <Button style="margin-right: 8px" @click="showRunTask = false">Cancel</Button>
        <Button type="primary" @click="onRunTask">RUN</Button>
      </div>
    </Drawer>
  </div>
</template>
<script>
import { Pipeline } from "@/api/pipeline.js";
import { Config } from "@/config";

export default {
  data() {
    return {
      page: 1,
      total: 0,
      pageSize: 10,
      columns: [
        {
          title: "ID",
          key: "id"
        },
        {
          title: "流水线名",
          key: "name"
        },
        {
          title: "备注",
          key: "desc"
        },
        {
          title: "创建时间",
          key: "created_at"
        },
        {
          title: "更新时间",
          key: "updated_at"
        },
        {
          title: "Action",
          key: "action",
          width: 260,
          align: "center",
          render: (h, params) => {
            return h("div", [
              h(
                "Poptip",
                {
                  props: {
                    confirm: true,
                    title: this.$t("public.confirmDelete")
                  },
                  style: {
                    display:
                      params.row.username == "admin" ? "none" : "inline-block"
                  },
                  on: {
                    "on-ok": () => {
                      this.delete(params.row);
                    }
                  }
                },
                [
                  h(
                    "Button",
                    {
                      props: {
                        type: "error",
                        size: "small"
                      },
                      style: {
                        marginRight: "5px"
                      }
                    },
                    this.$t("public.delete")
                  )
                ]
              ),
              h(
                "Button",
                {
                  props: {
                    type: "primary",
                    size: "small"
                  },
                  style: {
                    marginRight: "5px"
                  },
                  on: {
                    click: () => {
                      this.modify(params.row);
                    }
                  }
                },
                this.$t("public.edit")
              ),
              h(
                "Button",
                {
                  props: {
                    type: "primary",
                    size: "small"
                  },
                  style: {
                    marginRight: "5px"
                  },
                  on: {
                    click: () => {
                      this.runTask(params.row);
                    }
                  }
                },
                "RUN"
              )
            ]);
          }
        }
      ],
      data: [],

      // run task 弹框
      showRunTask: false,
      styles: {
        height: "calc(100% - 55px)",
        overflow: "auto",
        paddingBottom: "53px",
        position: "static"
      },
      runTaskInfo: {}, // 当前要运行的流水线
      taskId: "", // 当前运行的任务id
      taskList: [], // 运行的接口列表
      isClose: false, // 是否已关闭
      taskLogs: [], // 执行日志
      run_split: 0.3 // 分屏
    };
  },
  methods: {
    // 获取列表数据
    getList() {
      this.data = [
        {
          id: 1,
          project_id: 1,
          name: "测试"
        }
      ];
    },
    // 切换页码
    changeListPage(page) {
      this.page = page;
      this.getList();
    },
    // 页面打小
    pageSizeChange(pageSize) {
      this.pageSize = pageSize;
      this.changeListPage(1);
    },

    // 运行一个测试流水线
    runTask(row) {
      this.runTaskInfo = row;
      this.showRunTask = true;
    },
    // 运行任务
    onRunTask() {
      if (this.isClose == true) {
        this.initWebSocket();
      }
      this.taskLogs = [];
      Pipeline.RunTask(this.runTaskInfo.id, this.runTaskInfo.project_id).then(
        response => {
          if (response.status == 200) {
            console.log(response);
            this.taskId = response.data.task_id;
            this.taskList = response.data.tasks;

            let wsData = {
              type: "consumer",
              payload: JSON.stringify({ task_id: this.taskId })
            };
            this.websocketsend(JSON.stringify(wsData));
          }
        }
      );
    },

    // 日志颜色类型
    getAleryType(item, key){
      if(item.type == 'log'){
        return key%2 ? 'info' : 'warning'
      }else if(item.type == 'task_state'){
        return item.expect_type
      }
    },

    /* WebSocket */
    initWebSocket() {
      let wsuri = Config.WsBaseUrl + "/v1/pipeline/ws";
      this.websock = new WebSocket(wsuri);
      this.websock.onopen = this.websocketonopen;
      this.websock.onerror = this.websocketonerror;
      this.websock.onmessage = this.websocketonmessage;
      this.websock.onclose = this.websocketclose;
    },
    websocketping() {
      if (this.isClose == true) {
        return;
      }
      console.log("ping");
      this.websocketsend(`{"type":"ping"}`);
      setTimeout(() => {
        this.websocketping();
      }, 1000);
    },
    websocketonopen() {
      console.log("WebSocket连接成功");
      // this.websocketping();
    },
    websocketonerror(e) {
      console.log("WebSocket连接发生错误");
      this.isClose = true;
    },
    websocketonmessage(e) {
      //数据接收
      console.log(e.data);
      let data = JSON.parse(e.data);
      this.taskLogs.push(data);
      // 滚动到底部
      var div = document.getElementById("new_message");
      div.scrollTop = div.scrollHeight;
      // api 是否达到预期消息
      if (data.type == "task_state") {
        data.expect_type = 'error'
        let apiMsg = JSON.parse(data.msg);
        data.msg = '未达到预期'
        this.taskList.forEach(val => {
          if (val.id == apiMsg.id) {
            console.log(apiMsg);
            
            this.$set(val, "is_expect", apiMsg.expect);
            if (apiMsg.expect == true) {
              data.expect_type = 'success'
              data.msg = '已达到预期'
            }
          }
        });
      }
    },

    websocketsend(agentData) {
      console.log(agentData);

      //数据发送
      this.websock.send(agentData);
    },

    websocketclose(e) {
      //关闭
      console.log("connection closed (" + e.code + ")");
      this.isClose = true;
    }

    /* end WebSocket */
  },
  mounted() {
    this.initWebSocket();
    this.getList();
  },
  destroyed: function() {
    this.websocketclose();
  }
};
</script>

<style lang="scss" scoped>
.pipeline {
  padding: 10px;
  max-height: calc(100vh - 100px);
  .page {
    margin-top: 10px;
    text-align: right;
  }
  .search {
    width: 100%;
    height: 42px;
    overflow: hidden;
    .search-form {
      float: left;
    }
    .btns {
      float: right;
    }
  }
}
.drawer-footer {
  width: 100%;
  position: absolute;
  bottom: 0;
  left: 0;
  border-top: 1px solid #e8e8e8;
  padding: 10px 16px;
  text-align: right;
  background: #fff;
}
.line-title {
  font-size: 14px;
  font-weight: bold;
}
.line-content {
  padding-left: 5px;
}
.split-pane {
  padding-left: 10px;
  padding-bottom: 20px;
  height: calc(100vh - 130px);
  overflow-y: scroll;
}
.drawer-body{
  height: 100%;
  padding:10px;
  border: 1px solid #ccc;
}
</style>