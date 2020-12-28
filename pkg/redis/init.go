package redis

import (
	"github.com/go-redis/redis/v7"
	red "github.com/gomodule/redigo/redis"
	"github.com/matchstalk/go-admin-core/cache"
	"github.com/matchstalk/redisqueue"
	"go-api/common/global"
	"go-api/tools"
	"go-api/tools/config"
	"sync"
	"time"
)

//var RedisCli red.Conn

var once sync.Once

//单例
func GetRedis() *cache.Redis {
	once.Do(func() {
		host := config.RedisConfig.Host
		port := config.RedisConfig.Port
		global.Redis = &cache.Redis{
			ConnectOption: &redis.Options{
				Addr: host + ":" + tools.IntToString(port),//主机名+冒号+端口，默认localhost:6379
				//连接信息
				Network:  "tcp",                  //网络类型，tcp or unix，默认tcp
				Password: "",                     //密码
				DB:       0,                      // redis数据库index

				//连接池容量及闲置连接数量
				PoolSize:     15, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
				MinIdleConns: 10, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

				//超时
				DialTimeout:  5 * time.Second, //连接建立超时时间，默认5秒。
				ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
				WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
				PoolTimeout:  4 * time.Second, //当所有连接
			},
			ConsumerOptions: &redisqueue.ConsumerOptions{
				VisibilityTimeout: 60 * time.Second,
				BlockingTimeout:   5 * time.Second,
				ReclaimInterval:   1 * time.Second,
				BufferSize:        100,
				Concurrency:       10,
			},
			ProducerOptions: &redisqueue.ProducerOptions{
				StreamMaxLength:      100,
				ApproximateMaxLength: true,
			},
		}
		err := global.Redis.Connect()
		if err != nil{
			panic("redis connect err")
		}
	})

	return global.Redis
}

//连接池
func GetRedisPool() *red.Pool{
	once.Do(func() {
		host := config.RedisConfig.Host
		port := config.RedisConfig.Port
		redisPassword := ""
		maxIdle := 1024    //最大空闲链接数
		maxActive := 0     //最大链接数， 0 表示没有限制，可以无限链接
		idleTimeout := 120 //连接超时

		redisDsn := host + ":" + tools.IntToString(port)
		setRedisDb := red.DialDatabase(0) //默认库
		setRedisPassword := red.DialPassword(redisPassword)

		global.RedisPool = &red.Pool{
			MaxIdle:     maxIdle,
			MaxActive:   maxActive,
			IdleTimeout: time.Duration(idleTimeout),
			Dial: func() (red.Conn, error) {
				return red.Dial(
					"tcp",
					redisDsn,
					red.DialReadTimeout(time.Duration(1000)*time.Millisecond),
					red.DialWriteTimeout(time.Duration(1000)*time.Millisecond),
					red.DialConnectTimeout(time.Duration(1000)*time.Millisecond),
					setRedisDb,
					setRedisPassword,
				)
			},
		}
	})

	return global.RedisPool

}

//redis操作
func RedisDo(args ...interface{}) (res interface{}, err error){
	client := global.Redis.GetClient()
	res, err =client.Do(args ...).Result()
	defer client.Close()
	return res,err
}

//连接池操作
func RedisPoolDo(commandName string, args ...interface{}) (res interface{}, err error){
	conn := global.RedisPool.Get()
	res,err = conn.Do(commandName, args ...)
	defer conn.Close()
	return res,err
}


func Setup()  {

	// 打开后，一直不关闭会不会有问题？
	// 可以考虑 golang pool redis
	GetRedis()
	GetRedisPool()
}
