# P2P Bitcoin Yol Haritasi

Bu dosya mevcut merkezi blockchain projesini adim adim P2P, gossip, consensus ve kalici storage kullanan daha gercekci bir Bitcoin benzeri sisteme cevirmek icin yazildi.

Su anki yapi egitim icin iyi bir baslangic:

- `wallet`: private/public key, address ve transaction imzalama.
- `wallet_server`: HTML arayuz ve wallet API.
- `block`: block, blockchain, transaction pool, mining, balance hesabi.
- `blockchain_server`: merkezi blockchain HTTP API.
- `utils`: ECDSA ve JSON yardimcilari.

Ama su anki yapi P2P degil. Wallet tek bir blockchain server'a istek atiyor. Node'lar birbirini tanimiyor, block/transaction yaymiyor, fork cozumu yapmiyor ve chain memory'de tutuluyor.

## Hedef Mimari

Son hedef su olmali:

```text
wallet_server
    |
    | HTTP
    v
local node API
    |
    | internal call
    v
node
    |-- blockchain
    |-- mempool
    |-- consensus
    |-- storage
    |-- p2p network
            |
            | TCP/WebSocket/libp2p
            v
        other nodes
```

Wallet artik merkezi `blockchain_server` yerine kendi bagli oldugu local/full node'a istek atmali. Node hem HTTP API sunmali hem de diger node'larla P2P konusmali.

## Yeni Klasor Yapisi

Mevcut yapinin uzerine su klasorleri eklemek mantikli:

```text
bitcoin/
  block/
  wallet/
  wallet_server/
  utils/

  node/
    node.go
    api.go
    config.go
    bootstrap.go

  protocol/
    message.go
    version.go
    inventory.go
    codec.go

  network/
    peer.go
    server.go
    dialer.go
    manager.go
    gossip.go

  mempool/
    mempool.go
    validation.go

  consensus/
    validation.go
    fork_choice.go
    difficulty.go
    reward.go

  storage/
    storage.go
    block_store.go
    chain_state.go
    utxo_store.go
    peer_store.go

  cmd/
    node/
      main.go
    wallet_server/
      main.go
```

Ilk asamada `cmd/` sart degil. Mevcut `blockchain_server` ve `wallet_server` klasorleriyle devam edebilirsin. Ama proje buyuyunce `cmd/node/main.go` ve `cmd/wallet_server/main.go` daha temiz olur.

## Hangi Dosyaya Ne Yazilacak?

### `block/blockchain.go`

Burada sadece domain modelleri ve blockchain mantigi kalmali.

Kalmasi gerekenler:

- `Block`
- `Blockchain`
- `Transaction`
- `NewBlock`
- `Hash`
- `MarshalJSON`
- `LastBlock`
- `CreateBlock`

Zamanla buradan cikarilmasi gerekenler:

- HTTP ile ilgili hicbir sey burada olmamali.
- P2P ile ilgili hicbir sey burada olmamali.
- Mining timer gibi server davranisi burada olmamali.

Eklenmesi gerekenler:

- deterministic block hash
- transaction ID
- Merkle root
- block header
- block height
- cumulative work

Ornek hedef block:

```go
type BlockHeader struct {
	Version       int
	PreviousHash string
	MerkleRoot    string
	Timestamp     int64
	Bits          uint32
	Nonce         uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
	Height       int
	Hash         string
}
```

### `wallet/wallet.go`

Wallet sadece key, address ve transaction imzalama isi yapmali.

Kalmasi gerekenler:

- private key olusturma
- public key olusturma
- blockchain address olusturma
- transaction signature uretme

Eklenmesi gerekenler:

- amount icin `float32` yerine `int64` satoshi kullanimi
- UTXO tabanli transaction imzalama
- birden fazla input imzalama

Wallet chain tutmamali. Wallet sadece node'a sorar:

- benim bakiyem ne?
- bana ait UTXO'lar ne?
- bu transaction'i yayinla.

### `wallet_server/wallet_server.go`

Bu kisim UI/API gateway olarak kalabilir.

Wallet server nereye istek atmali?

Ilk asamada:

```text
wallet_server -> local node HTTP API
```

Yani `wallet_server` su adrese gider:

```text
http://127.0.0.1:5005
```

Ama bu adres artik merkezi blockchain server degil, local node API olmali.

Wallet server endpointleri:

```text
POST /wallet
GET  /wallet/amount?blockchain_address=...
POST /transaction
```

Wallet server'in node'a atacagi istekler:

```text
GET  {node}/amount?blockchain_address=...
POST {node}/transactions
GET  {node}/utxos?address=...
```

Gelecekte transaction olusturma akisi soyle olmali:

```text
1. wallet_server node'dan UTXO listesi ister.
2. wallet yeterli UTXO secimi yapar.
3. wallet transaction input/output olusturur.
4. wallet inputlari imzalar.
5. wallet_server signed transaction'i node'a POST eder.
6. node transaction'i validate eder.
7. node mempool'a alir.
8. node transaction'i gossip ile peer'lara yayar.
```

### `wallet_server/templates/index.html`

UI burada kalabilir.

Eklenmesi gerekenler:

- balance reload zaten var.
- transaction history listesi.
- pending transaction listesi.
- connected node bilgisi.
- node sync status bilgisi.

Ornek UI alanlari:

```text
Wallet
  public key
  private key
  address
  confirmed balance
  pending balance

Network
  connected node
  node height
  peers count
  sync status

Send
  recipient
  amount
  fee
```

### `blockchain_server/`

Kisa vadede bunu kullanmaya devam edebilirsin. Ama uzun vadede bu klasorun adi `node_server` veya `node` olmali.

Mevcut durumda:

```text
blockchain_server = merkezi chain API
```

Hedef durumda:

```text
node = chain API + P2P network + consensus + mempool + storage
```

Bu yuzden yeni gelistirmeleri direkt `blockchain_server` icine doldurmak yerine `node/`, `network/`, `consensus/`, `storage/` paketlerine bolmek daha dogru.

## Yeni Paketler

### `node/node.go`

Tum sistemi bir araya getiren ana struct burada olmali.

```go
type Node struct {
	ID         string
	HTTPAddr   string
	P2PAddr    string
	Blockchain *block.Blockchain
	Mempool    *mempool.Mempool
	Consensus  *consensus.Engine
	Network    *network.Manager
	Storage    storage.Storage
}
```

Node'un sorumluluklari:

- blockchain'i baslatmak
- storage'dan chain'i yuklemek
- mempool'u baslatmak
- P2P server'i baslatmak
- HTTP API'yi baslatmak
- gelen transaction'i validate edip mempool'a koymak
- gelen block'u validate edip chain'e baglamak
- gerekli mesajlari gossip ile yaymak

### `node/api.go`

Wallet ve debug icin HTTP endpointleri burada olmali.

Endpointler:

```text
GET  /chain
GET  /blocks/{hash}
GET  /headers
GET  /transactions
POST /transactions
GET  /amount?blockchain_address=...
GET  /utxos?address=...
POST /mine
GET  /peers
GET  /status
```

Wallet server bu API'ye istek atmali.

### `node/config.go`

Node ayarlari burada olmali.

```go
type Config struct {
	NodeID      string
	HTTPAddr    string
	P2PAddr     string
	DataDir     string
	Bootstrap   []string
	MinerAddress string
}
```

Ornek calistirma:

```powershell
go run ./cmd/node -http :5005 -p2p :6005 -data ./data/node1
go run ./cmd/node -http :5006 -p2p :6006 -data ./data/node2 -bootstrap 127.0.0.1:6005
go run ./cmd/node -http :5007 -p2p :6007 -data ./data/node3 -bootstrap 127.0.0.1:6005
```

### `protocol/message.go`

P2P mesaj formatlari burada olmali.

Baslangic icin JSON yeterli:

```go
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
```

Mesaj tipleri:

```text
version
verack
ping
pong
getaddr
addr
inv
getdata
tx
block
getheaders
headers
```

### `protocol/inventory.go`

Bitcoin'deki `inv` mantigi burada olmali.

```go
type InventoryType string

const (
	InvTx    InventoryType = "tx"
	InvBlock InventoryType = "block"
)

type InventoryItem struct {
	Type InventoryType `json:"type"`
	Hash string        `json:"hash"`
}
```

Gossip akisi:

```text
Node A transaction alir.
Node A peer'lara inv(tx_hash) yollar.
Peer'da tx yoksa getdata(tx_hash) ister.
Node A tx mesajiyla transaction'i yollar.
Peer validate eder, mempool'a koyar, baska peer'lara yayar.
```

### `network/server.go`

P2P listener burada olmali.

Sorumluluklari:

- TCP/WebSocket portunu dinlemek
- yeni peer kabul etmek
- peer handshake baslatmak
- peer mesajlarini okumak

Basit baslangic:

```go
func (s *Server) Start() error
func (s *Server) handleConn(conn net.Conn)
```

### `network/peer.go`

Tek bir peer baglantisi burada temsil edilmeli.

```go
type Peer struct {
	ID      string
	Addr    string
	Conn    net.Conn
	Inbound bool
	Send    chan protocol.Message
}
```

Peer sorumluluklari:

- mesaj okuma
- mesaj yazma
- ping/pong
- kapaninca manager'a haber verme

### `network/manager.go`

Peer listesi ve peer secimi burada olmali.

Sorumluluklari:

- connected peers
- known peers
- max inbound/outbound limit
- bootstrap peers
- reconnect
- misbehavior score

### `network/gossip.go`

Transaction ve block yayma burada olmali.

```go
func (m *Manager) BroadcastTransaction(tx *block.Transaction, exceptPeer string)
func (m *Manager) BroadcastBlock(b *block.Block, exceptPeer string)
func (m *Manager) BroadcastInventory(items []protocol.InventoryItem, exceptPeer string)
```

Sonsuz donguyu engellemek icin:

```go
seenTx     map[string]time.Time
seenBlocks map[string]time.Time
```

### `mempool/mempool.go`

Unconfirmed transaction'lar burada tutulmali.

```go
type Mempool struct {
	transactions map[string]*block.Transaction
}
```

Sorumluluklari:

- add transaction
- remove transaction
- transaction var mi kontrolu
- block'a giren transaction'lari silme
- miner icin transaction secme

### `mempool/validation.go`

Mempool'a girecek transaction kurallari burada olmali.

Kontroller:

- transaction ID dogru mu?
- imza gecerli mi?
- input UTXO var mi?
- input daha once harcanmis mi?
- toplam input >= toplam output + fee mi?
- duplicate transaction mi?

### `consensus/validation.go`

Block ve chain validasyonu burada olmali.

Kontroller:

- previous hash biliniyor mu?
- block hash target altinda mi?
- Merkle root dogru mu?
- timestamp kabul edilebilir mi?
- transaction'lar valid mi?
- coinbase transaction dogru mu?
- block reward dogru mu?
- double spend var mi?

### `consensus/fork_choice.go`

Fork cozumu burada olmali.

Kural:

```text
En uzun chain degil, en cok cumulative work iceren chain secilmeli.
```

Gerekli islemler:

- yeni block side chain'e eklenebilir.
- yeni branch daha cok work'e sahipse reorg yapilir.
- eski active chain'den ayrilan transaction'lar mempool'a geri alinabilir.
- yeni active chain UTXO set'e uygulanir.

### `consensus/difficulty.go`

Difficulty ayarlamasi burada olmali.

Baslangicta sabit difficulty kullanabilirsin. Sonra Bitcoin benzeri retarget eklenir.

Basit hedef:

```text
Her N block'ta bir difficulty ayarla.
Blocklar hedef sureden hizli geliyorsa difficulty artar.
Yavas geliyorsa difficulty azalir.
```

### `consensus/reward.go`

Mining reward burada olmali.

Kurallar:

- coinbase reward
- transaction fee toplamı
- halving
- max supply
- coinbase maturity

Basit baslangic:

```go
func BlockSubsidy(height int) int64
func CoinbaseMaturity() int
```

### `storage/storage.go`

Storage interface burada olmali.

```go
type Storage interface {
	SaveBlock(block *block.Block) error
	GetBlock(hash string) (*block.Block, error)
	SaveTip(hash string) error
	GetTip() (string, error)
	Close() error
}
```

Baslangic icin memory storage yazilabilir. Sonra BoltDB/BadgerDB/LevelDB'e gecilir.

### `storage/block_store.go`

Blocklar burada tutulmali.

Disk uzerinde:

```text
data/node1/blocks.db
```

Kayitlar:

```text
block:{hash} -> serialized block
height:{height} -> block hash
tip -> active chain tip hash
```

### `storage/chain_state.go`

Active chain bilgisi burada tutulmali.

Tutulacak seyler:

- active tip
- height index
- block metadata
- cumulative work

### `storage/utxo_store.go`

UTXO set burada tutulmali.

Kayit mantigi:

```text
utxo:{txid}:{index} -> output
```

Bir block active chain'e eklenince:

```text
1. input'lar UTXO set'ten silinir.
2. output'lar UTXO set'e eklenir.
```

Reorg olursa:

```text
1. eski branch geri alinir.
2. yeni branch uygulanir.
```

### `storage/peer_store.go`

Bilinen peer'lar burada tutulmali.

```text
peer:{addr} -> peer metadata
```

Node restart edince eski peer'lara tekrar baglanabilir.

## Zincir Nerede Tutulacak?

Kisa vadede:

```text
Blockchain struct icinde memory'de
```

Mevcut:

```go
type Blockchain struct {
	transactionPool []*Transaction
	chain []*Block
}
```

Bu sadece egitim icin yeterli. Server kapaninca chain kaybolur.

Orta vadede:

```text
storage/block_store.go
storage/chain_state.go
storage/utxo_store.go
```

Disk path:

```text
bitcoin/data/node1/
bitcoin/data/node2/
bitcoin/data/node3/
```

Her node kendi chain'ini kendi data klasorunde tutmali:

```text
data/node1/blocks.db
data/node1/chainstate.db
data/node1/peers.db

data/node2/blocks.db
data/node2/chainstate.db
data/node2/peers.db
```

Yani tum node'lar ayni dosyayi paylasmamali. Distributed sistemde her node kendi local state'ine sahiptir.

## Wallet Nereye Istek Atacak?

Wallet asla dogrudan baska peer'lara gossip mesaji atmamalı. Wallet su sekilde calismali:

```text
wallet_server -> local node HTTP API -> P2P network
```

Ornek:

```text
wallet_server gateway = http://127.0.0.1:5005
```

Wallet server su endpointleri kullanmali:

```text
GET  http://127.0.0.1:5005/amount?blockchain_address=...
GET  http://127.0.0.1:5005/utxos?address=...
POST http://127.0.0.1:5005/transactions
```

Node transaction'i aldiktan sonra:

```text
1. imzayi kontrol eder.
2. UTXO'lari kontrol eder.
3. mempool'a ekler.
4. peer'lara inv(tx_hash) yollar.
```

## Ilk P2P Versiyon Icin Minimum Endpointler

HTTP API:

```text
GET  /status
GET  /chain
GET  /amount?blockchain_address=...
GET  /transactions
POST /transactions
POST /mine
GET  /peers
POST /peers
```

P2P mesajlari:

```text
version
verack
ping
pong
inv
getdata
tx
block
```

Ilk versiyon icin `getheaders` ve `headers` sonraya birakilabilir.

## Ilk Calisan P2P Akisi

Uc node calistir:

```powershell
go run ./cmd/node -http :5005 -p2p :6005 -data ./data/node1
go run ./cmd/node -http :5006 -p2p :6006 -data ./data/node2 -bootstrap 127.0.0.1:6005
go run ./cmd/node -http :5007 -p2p :6007 -data ./data/node3 -bootstrap 127.0.0.1:6005
```

Wallet server node1'e baglansin:

```powershell
go run ./wallet_server -port 8080 -gateway http://127.0.0.1:5005
```

Beklenen akış:

```text
1. Wallet transaction olusturur.
2. Wallet transaction'i node1'e yollar.
3. Node1 transaction'i validate eder.
4. Node1 mempool'a ekler.
5. Node1 node2 ve node3'e inv(tx_hash) yollar.
6. Node2 ve node3 getdata ister.
7. Node1 tx mesajiyla transaction'i yollar.
8. Node2 ve node3 validate edip mempool'a ekler.
9. Node2 mine yaparsa block olusur.
10. Node2 inv(block_hash) yollar.
11. Node1 ve node3 block'u ister.
12. Node1 ve node3 block'u validate edip chain'e ekler.
13. Her node ayni height ve tip hash'e gelir.
```

## UTXO'ya Gecis Plani

Mevcut transaction:

```go
type Transaction struct {
	senderBlockchainAddress string
	recipientBlockchainAddress string
	value float32
}
```

Hedef transaction:

```go
type Transaction struct {
	ID      string
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	TxID      string
	OutIndex int
	Signature string
	PublicKey string
}

type TxOutput struct {
	Value   int64
	Address string
}
```

Balance hesabi artik chain'i bastan gezerek degil, UTXO set uzerinden yapilmali:

```text
balance(address) = address'e ait harcanmamis output toplamı
```

## Mining Nasil Olmali?

Mining akisi:

```text
1. mempool'dan valid transaction'lar secilir.
2. fee toplamı hesaplanir.
3. coinbase transaction olusturulur.
4. Merkle root hesaplanir.
5. block header hazirlanir.
6. nonce denenir.
7. target altinda hash bulunursa block uretilir.
8. block local chain'e uygulanir.
9. block peer'lara gossip ile yayilir.
```

Reward:

```text
reward = block subsidy + transaction fees
```

Para icin `float32` kullanilmamali. Hedef:

```go
type Satoshi int64
```

## Consensus Minimum Kurallari

Block kabul etmek icin:

- previous hash biliniyor olmali.
- block hash target altinda olmali.
- Merkle root dogru olmali.
- coinbase sadece ilk transaction olmali.
- coinbase reward fazla olmamali.
- tum transaction imzalari dogru olmali.
- double spend olmamali.
- UTXO'lar mevcut olmali.
- block timestamp cok ileri olmamali.

Fork choice:

```text
En fazla cumulative work'e sahip chain active chain olur.
```

## Sira Sira Yapilacaklar

### Asama 1: Mevcut kodu temizle

- `float32` yerine `int64` satoshi kullan.
- transaction ID ekle.
- block hash'i deterministic hale getir.
- Merkle root ekle.
- `blockchain_server.go.8557530553317608654` gibi gereksiz backup dosyalarini temizle.

### Asama 2: Storage ekle

- `storage/` klasorunu ac.
- once memory storage yaz.
- sonra BoltDB veya BadgerDB'e gec.
- block, tip, UTXO, peer bilgilerini disk'e yaz.

### Asama 3: Mempool ekle

- `mempool/` klasorunu ac.
- transaction pool'u `blockchain.go` icinden ayir.
- duplicate ve double spend kontrollerini ekle.

### Asama 4: Consensus ekle

- `consensus/` klasorunu ac.
- block validation ekle.
- fork choice ekle.
- chain reorg ekle.

### Asama 5: Node paketi ekle

- `node/` klasorunu ac.
- `Node` struct yaz.
- HTTP API'leri `node/api.go` icine tasi.
- `blockchain_server` uzun vadede node API'ye donussun.

### Asama 6: P2P protocol ekle

- `protocol/` klasorunu ac.
- mesaj tiplerini yaz.
- JSON codec ile basla.

### Asama 7: Network ekle

- `network/` klasorunu ac.
- peer manager yaz.
- TCP veya WebSocket server yaz.
- handshake, ping/pong ekle.

### Asama 8: Gossip ekle

- transaction gossip.
- block gossip.
- `inv/getdata/tx/block` akisini uygula.
- seen cache ekle.

### Asama 9: Multi-node test

- 3 node ayni makinede calissin.
- transaction node'lar arasinda yayilsin.
- mining sonrasi block tum node'lara gelsin.
- fork durumunda dogru chain secilsin.

### Asama 10: Guvenlik ve kalite

- rate limit.
- peer ban.
- malformed message reject.
- max message size.
- integration test.
- deterministic test network.

## Onemli Tasarim Kararlari

1. Wallet chain tutmaz.
2. Her node kendi chain'ini kendi data klasorunde tutar.
3. Wallet sadece local node HTTP API'ye istek atar.
4. Node diger node'larla P2P konusur.
5. Balance UTXO set'ten hesaplanir.
6. Consensus kurallari tum node'larda ayni ve deterministik olmalidir.
7. Mempool policy ile consensus ayridir.
8. Float para icin kullanilmaz.
9. Chain secimi cumulative work ile yapilir.
10. Gossip sonsuz donguye girmemesi icin seen cache kullanir.

## Okuma Sirasi

1. Bitcoin whitepaper.
2. Bitcoin Developer Guide - Block Chain.
3. Bitcoin P2P Network Reference.
4. UTXO modelini anlatan kaynaklar.
5. Kademlia paper.
6. Erlay paper.
7. BIP 152 Compact Block Relay.

Bu sirayla gidersen once consensus ve transaction temelini, sonra P2P/gossip tarafini daha saglam anlarsin.
