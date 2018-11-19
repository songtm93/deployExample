# 简介
Jenkins 是一个可扩展的持续集成引擎，主要用于持续、自动地构建/测试软件项目，还可用于监控一些定时执行的任务。本文将介绍如何在滴滴云上，使用Jenkins作为持续集成服务器，Git仓库作为代码存储库，实现工程的自动构建、部署等过程。

# 安装
以下均基于滴滴云CentOS 7.4镜像进行安装。
```
yum install -y java-1.8.0-openjdk   //首先安装jdk1.8或以上版本
sudo wget -O /etc/yum.repos.d/jenkins.repo http://jenkins-ci.org/redhat/jenkins.repo  //需要添加jenkins的yum源
sudo rpm --import http://pkg.jenkins-ci.org/redhat/jenkins-ci.org.key
yum install -y jenkins               //安装jenkins
```

# 运行
执行以下命令，启动jenkins进程
```
systemctl start jenkins
```
服务启动后，查看`/var/log/jenkins/jenkins.log`得到管理员密码。
```
Jenkins initial setup is required. An admin user has been created and a password generated.
Please use the following password to proceed to installation:

9811815bdf454944aea7a505de2c1d30

This may also be found at: /var/lib/jenkins/secrets/initialAdminPassword //也可以在这个目录下找到管理员密码
```
Jenkins默认监听8080端口，也可通过修改配置`vim /etc/sysconfig/jenkins`更换。
通过访问`http://公网IP:8080`来访问。
这里出现了无法访问的情况：`netstat -anp | grep 8080`查看进程存在，原来是被安全组拦住了，在滴滴云界面上添加一条安全组规则允许8080端口访问。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addSgRule.jpg" width="800"/>

再次访问，可以登录，输入之前获得的管理员密码进入jenkins的web页面。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/installPlugins.jpg" width="800"/>

安装插件、设置用户信息、设置Jenkins URL后，终于登陆到jenkins主界面。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/mainSurface.jpg" width="800"/>

安装完成后，接下来开始自动部署的配置。

# 使用webhook方式实现自动部署

## 生成Github Access Token
首先，为了使jenkins能够正确访问Github的API，在Github页面上点击`Setting ->Developer Settings -> Personal Access Token -> Generate new token`
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/githubSettingDeveloperSettings.jpg" width="800"/>
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/newToken.jpg" width="800"/>

选择`repo`和`admin:repo_hook`两种权限，第一个是为了正确访问github仓库，第二个是为了使用github的hook机制来实现对代码更新的感知。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/newTokenExample.jpg" width="800"/>

生成好的Token如图所示，此Token只能在生成时看到，如果关闭页面就无法再看到了，所以要谨慎保存。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/tokenExample.jpg" width="800"/>

## 配置jenkins的Github插件
对于jenkins的Github plugin，之前在初始化jenkins进入界面之前一般会默认勾选，并自动安装这个插件。如果没有安装也可以在jenkins的`系统管理->插件管理->可选插件`中找到`Github plugin`这个插件。

安装完毕后，需要在`系统管理->系统设置->GitHub`选项卡中添加一个Github Server，点击`Add GitHub Server`
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addGitServerButton.jpg" width="800"/>
选择添加凭据，选择Secret Text类型，并把刚才在Github上获取的Access Token填入，点击添加。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addSecretText.jpg" width="800"/>
添加完毕后，在Git Server一栏选择刚才添加的Secret Text，然后点击连接测试，提示`Credentials verified for user xxx, rate limit 4998`。表示连接成功。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/connectionTest.jpg" width="800"/>

## 添加构建任务
在Jenkins界面上，选择新建任务，选择构建一个自由风格的软件项目。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/createProject.jpg" width="800"/>

勾选Github项目，填入需要自动构建的github项目地址。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/gitURL.jpg" width="800"/>

在源码管理一项，勾选git。填入项目地址，并在`Credentials`项添加一个有访问权限的账户。在`源码库浏览器`一项选择githubweb，并依旧填入项目地址。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/codeManage.jpg" width="800"/>

为了实现push自动构建，还需要在构建触发器中选择`GitHub hook trigger for GITScm polling`，并在构建环境中选择`Use secret text(s) or file(s)`添加之前编辑的`Secret Text（Access Token）`。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/buildTrigger.jpg" width="800"/>

接下来编辑具体项目的构建命令。我在这里使用的示例代码是基于Go语言的，向页面发起请求会返回当前ip以及版本号。
```
export GOPATH=/home/dc2-user/gopath     //设置GOPATH
rm -rf $GOPATH/src/github.com/songtm93/deployExample
ln -s $JENKINS_HOME/workspace/deployExample $GOPATH/src/github.com/songtm93/deployExample   //链接
cd $GOPATH/src/github.com/songtm93/deployExample
rm -rf ./output/*   //删除上次编译输出
go build -v -o output/deployExample main.go //编译
make stop   //停止上次进程
./output/deployExample &    //运行
```
选择`构建步骤->执行shell`，编辑构建命令。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/buildCommand.jpg" width="800"/>

Jenkins的工作空间为`$JENKINS_HOME`，默认为`/var/lib/jenkins`，可以在`系统管理->系统设置->主目录`下看到。
Jenkins储存所有的数据文件在这个目录下，可以通过设置`$JENKINS_HOME`环境变量或者更改`Jenkins.war`内的`web.xml`设置文件来修改。这个值在Jenkins运行时是不能更改的，这里不做修改，将jenkins pull下来的代码作软链到`$GOPATH`目录下确保程序能够正常编译。

## 执行构建
至此一个简单的自动构建项目已经配置完毕。点击自动构建进行第一次构建。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/firstBuild.jpg" width="800"/>

这里普遍会遇到一个问题，在jenkins的构建过程中，使用shell启动web服务进程后，在构建任务完成时，jenkins会把构建中启动的所有子进程杀掉，导致web进程也down掉了。这个问题可以通过在运行web进程的命令前临时更改构建任务的`$BUILD_ID`来解决。
```
OLD_BUILD_ID=$BUILD_ID  //记录当前BUILD_ID
BUILD_ID=DONTKILLME     //更改当前BUILD_ID
make run                //启动web进程
BUILD_ID=$OLD_BUILD_ID  //改回BUILD_ID
```

提示构建成功后，访问对应的滴滴云主机Ip+端口，终于得到正确的返回值。
```
{"ip":"10.255.20.171","version":"1.0"}
```

更改当前版本号为"2.0"，并git push，成功触发了自动构建。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/autoTriggerBuild.jpg" width="800"/>

# 使用Publish Over Ssh Plugin实现批量自动部署
## 安装Publish Over Ssh Plugin 插件
在jenkins执行构建完毕后，可以使用插件实现批量自动部署到多台服务器。首先，在`系统管理->插件管理->可选插件` 中找到`Publish Over SSH` 插件直接安装并重启Jenkins。

## 配置Publish Over SSH
首先在控制台创建两台DC2，作为代码发布的目标节点，由于是内网访问，可以使用内网IP，以之前创建的那台DC2为跳板，因此不创建EIP实例。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/createSlave.jpg" width="800"/>

然后在`系统管理->系统设置`中找到`Publish Over SSH`的配置项，选择添加SSH Server。这里`Passphrase`为目标服务器的登录密码，`Path to key`和`key`为可用于登录目标服务器的sshkey私钥。但要注意在密码和私钥两项中，如果填了私钥就会默认优先使用私钥登录，若没有事先拷贝公钥到目标服务器，则会一直导致登录失败，笔者就遇到了这个问题。

配置完毕后，点击`Test Configuration`可以测试是否可以连接目标服务器。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addSlave.jpg" width="800"/>

另外，也可以在具体的SSH Server下添加对应其的密码或私钥。注意`Remote Directory`里必须填此用户有权限的目录。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addSlave2.jpg" width="800"/>

## 配置自动构建任务
更改好配置后，回到刚才的构建任务，`更改配置->构建后操作->增加构建后操作步骤->Send build artifacts over SSH`
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addPostBuildStep1.jpg" width="800"/>

这里选择刚才添加的目标服务器，并编辑配置。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/addPostBuildStep2.jpg" width="800"/>

## 运行
配置好服务后，之后就可以验证自动部署的集群是否可以正常工作了。使用SLB能够方便的进行节点间的负载均衡，在控制台创建一个SLB实例。
<img src="https://raw.githubusercontent.com/songtm93/deployExample/master/pics/createSLB.jpg" width="800"/>

创建好SLB之后，多次请求SLB的EIP，看是否能够将请求分发到不同服务器上，以及不同服务器是否均部署完毕：
```
curl http://117.51.157.199/
{"ip":"10.255.20.171","version":"2.0"}
curl http://117.51.157.199/
{"ip":"10.255.20.47","version":"2.0"}
curl http://117.51.157.199/
{"ip":"10.255.20.154","version":"2.0"}
curl http://117.51.157.199/
{"ip":"10.255.20.171","version":"2.0"}
curl http://117.51.157.199/
{"ip":"10.255.20.47","version":"2.0"}
curl http://117.51.157.199/
{"ip":"10.255.20.154","version":"2.0"}
```
证明集群的自动部署均已成功。

# 参考
1. [jenkins使用Publish Over SSH插件实现远程自动部署](http://blog.51cto.com/xiong51/2091739)

2. [Jenkins linux 操作系统一键部署多节点](https://blog.csdn.net/erbao_2014/article/details/62430518)

3. [构建基于Jenkins + Github的持续集成环境](https://blog.csdn.net/it_hue/article/details/79353563)

4. [实战：向GitHub提交代码时触发Jenkins自动构建](https://blog.csdn.net/boling_cavalry/article/details/78943061)

5. [解决 jenkins 自动杀掉进程大坑](https://blog.csdn.net/recotone/article/details/80510201?utm_source=blogxgwz8)