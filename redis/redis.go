package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mediocregopher/radix/v3"
)

var ctx = context.Background()

// ClientOptions struct contains the options for connecting to redis
type ClientOptions struct {
	Host            string
	Port            string
	Password        string
	MaxRetries      int
	MinRetryBackOff time.Duration
	MaxRetryBackOff time.Duration
	WriteTimeout    time.Duration
	DB              int
	PoolSize        int
}

var dummyHashMap = make(map[string]string)

// Client struct holds connection to redis
type Client struct {
	conn *redis.Client
}

// Clientv2 struct holds pool connection to redis using radix dep
type Clientv2 struct {
	pool *radix.Pool
}

// NewClient method will return a pointer to new client object
func NewClient(opts *ClientOptions) *Client {
	var poolSize = 20
	if opts.PoolSize > 0 {
		poolSize = opts.PoolSize
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:            opts.Host + ":" + opts.Port,
		Password:        opts.Password,
		DB:              opts.DB,
		MaxRetries:      opts.MaxRetries,
		MinRetryBackoff: opts.MinRetryBackOff,
		MaxRetryBackoff: opts.MaxRetryBackOff,
		WriteTimeout:    opts.WriteTimeout,
		PoolSize:        poolSize,
	})
	var client = &Client{conn: redisClient}
	return client
}

// NewV2Client will return the pool connection to radix object
func NewV2Client(opts *ClientOptions) *Clientv2 {
	// Ref: https://github.com/mediocregopher/radix/blob/master/radix.go#L107
	customConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(opts.WriteTimeout),
			radix.DialAuthPass(opts.Password),
			radix.DialSelectDB(opts.DB),
		)
	}
	poolSize := opts.PoolSize
	if poolSize == 0 {
		poolSize = 15
	}

	rclient, _ := radix.NewPool("tcp", opts.Host+":"+opts.Port, poolSize, radix.PoolConnFunc(customConnFunc))
	var client = &Clientv2{pool: rclient}
	return client
}

// GetConn returns a pointer to the underlying redis library
func (c *Client) GetConn() *redis.Client {
	return c.conn
}

// HIncrBy will increment a hash map key
func (c *Client) HIncrBy(key, field string, inc int64) int64 {
	resp := c.conn.HIncrBy(ctx, key, field, inc)
	result, _ := resp.Result()
	return result
}

// HIncrByFloat will increment a hash map key
func (c *Client) HIncrByFloat(key, field string, inc float64) float64 {
	resp := c.conn.HIncrByFloat(ctx, key, field, inc)
	result, _ := resp.Result()
	return result
}

// HGetAll will return the hash map
func (c *Client) HGetAll(key string) map[string]string {
	resp := c.conn.HGetAll(ctx, key)
	result, err := resp.Result()
	if err != nil {
		return dummyHashMap
	}
	return result
}

// HGet will return the hash map key
func (c *Client) HGet(key, field string) (string, bool) {
	resp := c.conn.HGet(ctx, key, field)
	result, err := resp.Result()
	if err != nil {
		return "", false
	}
	return result, true
}

// Get will return the hash map key
func (c *Client) HSet(key string, field string, value string) {
	c.conn.HSet(ctx, key, field, value)
}

func (c *Client) HDel(key string, field string) {
	c.conn.HDel(ctx, key, field)
}

func (c *Client) Hlen(key string) int64 {
	resp := c.conn.HLen(ctx, key)
	result, err := resp.Result()
	if err != nil {
		return 0
	}
	return result
}

func (c *Client) Hmset(key string, hashdata map[string]interface{}) {
	c.conn.HMSet(ctx, key, hashdata)
}

// Get will return the hash map key
func (c *Client) Get(key string) (string, bool) {
	resp := c.conn.Get(ctx, key)
	result, err := resp.Result()
	if err != nil {
		return "", false
	}
	return result, true
}

// Set will create a key in redis
func (c *Client) Set(key, field string, expiration time.Duration) {
	c.conn.Set(ctx, key, field, expiration)
}

func (c *Client) Keys(pattern string) *redis.StringSliceCmd {
	result := c.conn.Keys(ctx, pattern)
	return result
}

// Del method will remove single key from redis
func (c *Client) Del(key string) {
	c.conn.Del(ctx, key)
}

// DelMulti method will remove multiple keys from redis
func (c *Client) DelMulti(keys []string) {
	c.conn.Del(ctx, keys...)
}

// HIncrBy will increment a hash map key
func (c *Clientv2) HIncrBy(key, field string, inc int64) {
	if c.pool != nil {
		val := strconv.Itoa(int(inc))
		c.pool.Do(radix.Cmd(nil, "HINCRBY", key, field, val))
	}
}

// HIncrByFloat will increment a hash map key
func (c *Clientv2) HIncrByFloat(key, field string, inc float64) {
	if c.pool != nil {
		val := fmt.Sprintf("%f", inc)
		c.pool.Do(radix.Cmd(nil, "HINCRBYFLOAT", key, field, val))
	}
}

// SCard will get the size of set
func (c *Clientv2) SCard(key string) int {
	var count int
	if c.pool != nil {
		c.pool.Do(radix.Cmd(&count, "SCARD", key))
	}
	return count
}

// SIsMember will will check if value is in the set
func (c *Clientv2) SIsMember(key, val string) int {
	var isMember int
	if c.pool != nil {
		c.pool.Do(radix.Cmd(&isMember, "SISMEMBER", key, val))
	}
	return isMember
}

// SAdd will add the member to the set
func (c *Clientv2) SAdd(key, field string) int {
	var success int
	if c.pool != nil {
		c.pool.Do(radix.Cmd(&success, "SADD", key, field))
	}
	return success
}

// Close method closes the redis connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Close method closes the redis connection
func (c *Clientv2) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}
