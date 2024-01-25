
#!/bin/bash

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
fi


echo "start $process_name"
nohup ./proxypool -c ./config_file/config.yaml &
