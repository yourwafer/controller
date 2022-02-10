### 服务器CD控制
1. 编译 go build -o manager,运行./manager --help查看配置，默认以agent方式启动
2. 在目标控制机器启动agent，./manager -type agent
3. 在控制节点配置 config.json
```
{
  "project": "shennu",
  "description": "神怒",
  "baseDir": "F:\\tmp",
  "svnUser": "deploy",
  "svnPass": "***",
  "branches": {
    "qa3": {
      "description": "繁体",
      "agent": "192.168.11.102:11000",
      "serverPort": 11111,
      "svn": {
        "game": {
          "path": "svn://192.168.11.200/shennu/server/trunk",
          "description": "游戏服"
        },
        "center": {
          "path": "svn://192.168.11.200/shennu/server/trunk",
          "description": "中央服"
        },
        "numerical": {
          "path": "svn://192.168.11.200/shennu/numerical/trunk",
          "description": "游戏服数值表"
        },
        "client": {
          "path": "svn://192.168.11.200/shennu/numerical/trunk",
          "description": "客户端",
          "show": true
        }
      },
      "mysql": {
        "username": "root",
        "password": "root",
        "address": "127.0.0.1:3306",
        "databases": {
          "qa3_game": "",
          "qa3_center": ""
        }
      },
      "configs": [
        {
          "fileName": "game/resources/server.properties",
          "values": {
            "-xa.socket.address": "0.0.0.0:11111",
            "-server.resource.path": "file:///F:\\\\tmp\\\\shennu\\\\qa3\\\\numerical"
          }
        },
        {
          "fileName": "game/resources/jdbc.properties",
          "values": {
            "-jdbc.url": "jdbc:mysql://127.0.0.1/qa3_game?useUnicode=true&characterEncoding=utf-8",
            "-jdbc.username": "root",
            "-jdbc.password": "root"
          }
        },
        {
          "fileName": "game/resources/zookeeper.properties",
          "values": {
            "-zookeeper.serverPath": "/shennu_qa3"
          }
        },
        {
          "fileName": "center/resources/center/server.properties",
          "values": {
            "-xa.socket.address": "0.0.0.0:11112",
            "-server.resource.path": "file:///F:\\\\tmp\\\\shennu\\\\qa3\\\\numerical"
          }
        },
        {
          "fileName": "center/resources/server.properties",
          "values": {
            "-xa.socket.address": "0.0.0.0:11112",
            "-server.resource.path": "file:///F:\\\\tmp\\\\shennu\\\\qa3\\\\numerical"
          }
        },
        {
          "fileName": "center/resources/center/jdbc.properties",
          "values": {
            "-jdbc.url": "jdbc:mysql://127.0.0.1/qa3_center?useUnicode=true&characterEncoding=utf-8",
            "-jdbc.username": "root",
            "-jdbc.password": "root"
          }
        },
        {
          "fileName": "center/resources/center/zookeeper.properties",
          "values": {
            "-zookeeper.serverPath": "/shennu_qa3"
          }
        }
      ],
      "java": [
        {
          "name": "center",
          "commands": {
            "start": {
              "javaClass": "com.xa.shennu.center.Start",
              "memory": "2G"
            },
            "stop": {
              "javaClass": "com.xa.shennu.game.console.command.Stop",
              "memory": "128M"
            },
            "reload": {
              "javaClass": "com.xa.shennu.game.console.command.Reload",
              "memory": "128M"
            }
          }
        },
        {
          "name": "game",
          "commands": {
            "start": {
              "javaClass": "com.xa.shennu.game.Start",
              "memory": "2G"
            },
            "stop": {
              "javaClass": "com.xa.shennu.game.console.command.Stop",
              "memory": "128M"
            },
            "reload": {
              "javaClass": "com.xa.shennu.game.console.command.Reload",
              "memory": "128M"
            }
          }
        }
      ]
    },
    "qa4": {
      "description": "繁体2",
      "agent": "192.168.11.102:11000"
    },
    "qa5": {
      "description": "繁体3",
      "agent": "192.168.11.102:11000"
    },
    "qa6": {
      "description": "繁体4",
      "agent": "192.168.11.102:11000"
    },
    "qa7": {
      "description": "繁体5",
      "agent": "192.168.11.102:11000"
    }
  }
}
```
4. 启动控制节点，./manager -type server
5. 访问 http://localhost:11000

### 支持功能
1. 修改properties文件
2. 启动java进程，关闭java进程
3. 创建数据库，删除数据库，执行数据脚本
4. svn检出、更新
5. 修改系统时间
6. 向游戏服务器发送指令