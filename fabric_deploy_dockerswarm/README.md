部署流程：
环境的搭建可参考https://wiki.dianrong.com/pages/viewpage.action?pageId=32546493
1.拉取最新代码，并拷贝至作为manager的主机上
2.在每个节点上拉取最新镜像，可参考https://wiki.dianrong.com/pages/viewpage.action?pageId=29887117
3.进入deploy目录，执行fabric_deploy.sh脚本，可自动生成相应的证书，并将证书拷贝至其他需要运行节点的主机上
  3.1.目前生成证书的binary(cryptogen，configtxgen）可在ubuntu16.04，go1.8环境下正常使用
  3.2.在copyArtifacts.sh中可以修改需要拷贝证书的节点，包括节点的数量，每个节点的ip，放置证书的位置（该位置修改后需要同步修改docker-compose文件中证书挂载的位置），每个节点上运行的服务名称
  3.3.该脚本通过scp拷贝证书，进行如下配置后manager可免密向worker节点拷贝文件
    在manager节点上执行如下命令来生成配对密钥：
    ssh-keygen -t rsa
    遇到提示回车默认即可，公钥被存到用户目录下.ssh目录：
    ~/.ssh/id_rsa.pub
    将id_rsa.pub中的内容复制到每一个worker节点的~/.ssh/authorized_keys文件中，如果没有则新建该文件
    完成以上步骤后，从manager上scp文件至worker就不需要密码了
4.设置环境变量：source setEnvVars.sh
5.在manager上执行docker stack deploy -c file-name.yml stack-name部署服务
6.启动完成后，执行docker service ls命令查看容器启动情况，并执行docker service logs + service-name来查看相应的log。
