
#!/bin/bash


# postgresql 相关操作

# sudo /etc/init.d/postgresql start   # 开启
# sudo /etc/init.d/postgresql stop    # 关闭
# sudo /etc/init.d/postgresql restart # 重启

# 确保数据库进程启动
process_name="postgres"
if pgrep -x "$process_name" > /dev/null
then
    # 进程存在，终止进程
    echo "$process_name alreay started."
else
    # 进程不存在
    echo "$process_name is not running"
    /etc/init.d/postgresql start
fi

# 定义进程名
process_name="proxypool"

# 检查进程是否存在
if pgrep -x "$process_name" > /dev/null
then
    # 进程存在，终止进程
    echo "Killing $process_name"
    pkill -x "$process_name"
else
    # 进程不存在
    echo "$process_name is not running"
    echo "start $process_name"
fi


echo "start $process_name"
nohup ./proxypool -c ./config_file/config.yaml &
