<h1 align="center">Welcome to soonsocks 👋</h1>
<p>
  <img alt="Version" src="https://img.shields.io/badge/version-1.0-blue.svg?cacheSeconds=2592000" />
  <a href="https://yangsoon.github.io/#/posts/39" target="_blank">
    <img alt="Documentation" src="https://img.shields.io/badge/documentation-yes-brightgreen.svg" />
  </a>
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>
</p>

> A simple shadowsocks proxy implemented in golang

<div align="center"><img src="./img/soonsocks.png"/></div>

## Install

```sh
go mod download
```

## Usage

```sh
cd cmd/ssserver && go build -o sserver && ./sserver -c ../../testdata/config.json
```

## About

![soonsocks](https://user-images.githubusercontent.com/29531394/71320953-fa18cd80-24ed-11ea-94fe-889613014c5f.png)

本文将介绍一个简化版本的shadowsocks— **[soonsocks](https://github.com/yangsoon/soonsocks)** 的实现，soonsocks源于腾讯大佬[cssivision](https://github.com/cssivision/shadowsocks)的项目并做了部分修改使得更便于阅读和理解，项目对shadowsocks-go的实现进行了简化，只支持三种加密算法以及只支持一台服务器提供代理服务。下面本文将从socks5协议开始讲解soonsocks的实现，帮助大家更好的了解shadowsocks。

如果您觉得文章不错，欢迎给项目一个star **[soonsocks](https://github.com/yangsoon/soonsocks)** 

### socks5协议

关于socks5协议，网上的讲解[^1]有很多，而且sock5协议的[RFC 1928](https://www.ietf.org/rfc/rfc1928.txt)很短，包格式也很少。下图展示了socks5客户端和代理采用无认证的方式(ss就是采用无认证的方式)进行连接并传输数据的流程。主要是为了和shadowsocks的实现做对比。

![socks5](https://user-images.githubusercontent.com/29531394/71319914-62ac7e00-24df-11ea-9d78-ca64a80aab92.png)

主要流程分为socks5协商、建立连接和传输阶段: 

在协商阶段，客户端向代理发送请求协商认证方式，代理端告诉客户端采用无认证的方式

在连接阶段，客户端向代理发送要请求的目的地址，代理端会响应客户端是否建立连接

SOCKS5 协议只负责建立连接，在完成握手阶段和建立连接之后，SOCKS5 服务器就只做简单的转发了。客户端就开始传输数据，代理端将来自remote的响应返回给客户端。

### shadowsocks如何使用socks5建立连接

socks5因为简单易用的特性被shadowsocks来实现代理功能，其实socks5只被用来建立连接，其核心是使用一系列加密算法加密数据以及使用代理服务器做转发来实现一些网站的访问。下图为shadowsocks进行网站访问的整体流程。

![shadowsocks](https://user-images.githubusercontent.com/29531394/71322079-20df0000-24fe-11ea-8595-cf191c171177.png)

和传统的socks5不同，shadowsocks将Proxy划分为SSLocal和SSServer两个部分，其中SSLocal部署在国内无法科学上网的主机上，SSServer部署在国外的主机上，其中SSLocal可以和SServer进行tcp连接。

主要流程也是分为协商连接和数据传输阶段，协商和连接阶段和传统的socks5协议一致，在数据传输阶段，SSLocal会像SSServer传输加密后的客户端请求访问的网站地址，SSServer接收到加密数据之后解密确定将要代理访问的目标地址。SSLocal发送完请求地址之后便开始接收来自客户端的请求数据加密转发给SSServer。开始数据传输。

### SSLocal模块实现

SSLocal模块的实现严格按照上面的连接流程。首先SSLocal模块读取配置文件信息，绑定配置文件中指定的端口，并监听来自客户端的请求，每接收到一个请求就开启一个goroutine处理该连接的数据请求和响应。

```go
func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.json", "json file with config")
	flag.Parse()

	var err error
	config, err = ss.ParseConfig(configPath)
	if err != nil {
		ss.Logger.Fatalf("parse %s failed %v \n", configPath, err)
	}
	ss.Logger.Printf("SSLocal is running at %v\n", config.LocalAddr)
	ss.Logger.Printf("config info: \n"+
		"--------------------------------\n"+
		"LocalAddr: %v\n"+
		"ServerAddr: %v\n"+
		"Method: %v\n"+
		"--------------------------------\n",
		config.LocalAddr,
		config.ServerAddr,
		config.Method)

	l, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		ss.Logger.Printf("SSLocal listen faild %v\n", err)
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			ss.Logger.Printf("SSLocal accept client error: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}
```

handleConnection函数实现这样的功能：首先完成和客户端的socks5连接，在`ss.HandleShake`中完成，建立连接之后，和SSServer建立tcp连接，并传输客户端想要访问的地址，之后SSLocal就充当一个转发者的角色，将来自客户端的请求加密转发给SSServer，同时将来自SSServer的响应请求解密转发给客户端。

```go
func handleConnection(conn net.Conn) {
	rawaddr, host, err := ss.HandleShake(conn)
	if err != nil {
		ss.Logger.Printf("socks negotiate host %s error: %v\n", host, err)
		return
	}

	cipher, err := ss.NewCipher(config.Method, config.Password)

	if err != nil {
		ss.Logger.Printf("create cipher error: %v\n", err)
		return
	}

	serverCConn, err := ss.DialWithCipher(config.ServerAddr, cipher.Clone())
	if err != nil {
		ss.Logger.Printf("connect to server %s error: %v\n", config.ServerAddr, err)
		return
	}

	ss.Logger.Printf("connecting to server %v (request host %v)\n", config.ServerAddr, host)
	_, err = serverCConn.Write(rawaddr)
	if err != nil {
		ss.Logger.Printf("write to server %s error: %v\n", config.ServerAddr, err)
	}

	go func() {
		defer conn.Close()
		_, err := ss.CopyBuffer(conn, serverCConn)
		if err != nil {
			ss.Logger.Printf("connecting to %v error: %v\n", host, err)
		}
	}()

	_, err = ss.CopyBuffer(serverCConn, conn)
	if err != nil {
		ss.Logger.Printf("connecting to %v error: %v", host, err)
	}
	serverCConn.Close()
}
```

函数HandleShake完成了客户端和SSLocal的连接的建立，分为4个步骤和socks5协商连接阶段一一对应。

```go
func HandleShake(conn net.Conn) (rawaddr []byte, host string, err error) {

	rawaddr = []byte{}
	host = ""

	// 1. get pkg from client
	if _, err = extractNegotiation(conn); err != nil {
		return
	}
	Logger.Println("get conn from client")

	// 2. reply to client build connect
	if err = replyNegotiation(conn); err != nil {
		return
	}
	Logger.Println("reply to client")

	// 3. get request pkg from client
	var socks5r Socks5Request
	if socks5r, err = extractRequest(conn); err != nil {
		return
	}
	Logger.Printf("request %s\n", socks5r.Host)

	// 4. reply to client
	if err = replyRequest(conn); err != nil {
		return
	}
	Logger.Println("reply to client request")

	rawaddr = socks5r.RawAddr
	host = socks5r.Host
	return
}
```

### SSServer模块实现

该模块的实现逻辑和SSLocal基本相同，可以直接看代码实现[^ 2] 。其中需要注意的是，我们知道在shadowsocks传输数据的时候，首先SSlocal会发送一个数据包，包括接下来client将要请求的目标地址，这个目标地址是裁剪自socks5协商连接阶段，client发送给proxy请求目标地址的数据包中的目标地址字段，因此，SSServer还要需要先对目标数据包做解析。

### 数据加密模块

soonsocks支持rc4md和aes-128-cfb和aes-256-cfb三种加密算法，本部分主要讲解代理是如何使用aes加密算法进行数据加密传输。其中对aes原理只是简单一提，具体的加密算法请自行搜索。

#### 高级加密标准 AES 

AES 密码学中的高级加密标准(Advanced Encryption Standard，AES)又称高级加密标准Rijndael加密法。加密算法分为对称加密和非对称加密，两者的区别在于加密和解密所用的密钥是否为同一个密钥，其中Rijndael加密法属于对称加密算法。

AES的基本要求是，采用对称分组密码体制，密钥长度的最少支持为128、192、256，分组长度必须为128比特，密钥长度可以是128比特、192比特、256比特中的任意一个（如果数据块及密钥长度不足时，会补齐）。aes-128-cfb和aes-256-cfb的不同大家也能看出来了，就是密钥的长度不同。

![AES](https://user-images.githubusercontent.com/29531394/72031935-ebd9f800-32c8-11ea-8111-d3535fccfc2e.png)


#### AES CFB模式

CFB的加密过程分成两部分，先将前一段加密得到的密文加密，然后加密后的结果和当前明文异或。在对第一个块进行加密使用的 IV 即初始向量（Initialization Vector）它的作用和MD5的“加盐”有些类似，目的是防止同样的明文块始终加密成同样的密文块。

![CFB](https://user-images.githubusercontent.com/29531394/72031949-f3999c80-32c8-11ea-84f0-0f97ef464d7f.png)


CFB的解密过程几乎就是颠倒的CBC的加密过程。

> 图中虽然画的是解密器，但实际上解密器进行的操作仍然是使用和加密过程一样的算法对密文做加密处理。通过图中的公式可以看到这一点。

![CFBde](https://user-images.githubusercontent.com/29531394/72031956-f8f6e700-32c8-11ea-8826-61114a061ab9.png)


关于AES加密部分的代码实现，golang官方也提供了样例。soonsock中加密解密的部分和下面代码类似，就不再赘述。

```go
func ExampleNewCFBEncrypter() {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	plaintext := []byte("some plaintext")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	fmt.Printf("%x\n", ciphertext)
}

func ExampleNewCFBDecrypter() {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	ciphertext, _ := hex.DecodeString("7dd015f06bec7f1b8f6559dad89f4131da62261786845100056b353194ad")

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	fmt.Printf("%s", ciphertext)
	// Output: some plaintext
}
```

#### SS加密解密传输流程

下图描述了SSLocal和SSServer之间进行数据加密传输的过程，首先SSServer按照配置项监听指定端口，SSLocal连接到SSServer得到了连接conn并将连接包装为对象NewConn，NewConn包含有net.Conn对象和Cipher对象，并且实现了net.Conn的Read和Write接口。当使用连接读取和写入数据的时候，使用Cipher进行解密和加密。

同理，当SSServer获取到来自SSLocal的连接的时候，也会将conn包装成NewConn，使用Cipher对数据进行加密和解密。

![encrypted](https://user-images.githubusercontent.com/29531394/71407632-8d7d0a80-2676-11ea-8a69-7cc535e7a47e.png)

SSLocal和SSServer建立连接的时候，SSLocal先向连接中写入数据，首次写入时，SConn会先生成初始向量iv并用来初始化加密器enc，在写入数据的时候，会对要发送的请求数据加密，并在加密数据前附加上初始向量iv，因为解密过程需要使用相同的iv进行解密，所以SSLocal会在数据包前附上初始向量iv。

SSServer监听到来自SSLocal的连接之后，同样会将conn包装成LConn，SSServer接收到包含初始向量iv的数据包之后，会使用iv来初始化自己的解密器，以便解密数据包。同样当SSServer第一次使用LConn向SSLocal写入数据的时候，也会初始化自己的加密器，将对应的初始向量iv写到数据包中。SSLocal获取到带有iv的数据包并发现自己没有相应的解密器，所以就使用iv初始化自己的解密器，解密数据包。

当SSLocal和SSServer都初始化了自己加密解密器之后，接下来发送的数据包都不需要携带初始向量了。

```go
func (cc *CConn) Read(b []byte) (int, error){
	if cc.dec == nil {
		iv := make([]byte, cc.info.ivLen)
		if _, err := io.ReadFull(cc.Conn, iv); err != nil {
			return 0, err
		}
		if err := cc.initDecrypt(iv); err!=nil {
			return 0, err
		}

		if len(cc.iv) == 0 {
			cc.iv = iv
		}
	}
	encryptData := make([]byte, len(b))
	n, err := cc.Conn.Read(encryptData)
	if n > 0 {
		cc.Decrypt(b[0:n], encryptData[0:n])
	}
	return n, err
}

func (cc *CConn) Write(b []byte) (int, error) {
	var iv []byte
	if cc.enc == nil {
		if err := cc.initEncrypt(); err != nil {
			return 0, err
		}
		if len(cc.iv) == 0 {
			return 0, errors.New("get iv error")
		}
		iv = cc.iv
	}

	encryptData := make([]byte, len(iv)+len(b))
	if len(iv) > 0 {
		copy(encryptData, iv)
	}

	cc.Encrypt(encryptData[len(iv):], b)
	n, err := cc.Conn.Write(encryptData)
	return n, err
}
```

[^1]: http://zhihan.me/network/2017/09/24/socks5-protocol/
[^ 2]: https://github.com/yangsoon/soonsocks/blob/master/cmd/ssserver/main.go



## Author

👤 **yangsoon**

* Website: yangsoon.github.io
* Github: [@yangsoon](https://github.com/yangsoon)

## 🤝 Contributing

Contributions, issues and feature requests are welcome!<br />Feel free to check [issues page](https://github.com/yangsoon/soonsocks/issues). 

## Show your support

Give a ⭐️ if this project helped you!

***
_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_
