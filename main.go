package main

import (
    "os"
    "log"
    "fmt"
    "time"
    "math/big"
    "net/http"
    "crypto/sha256"
    "html/template"
    "path/filepath"
    rand "math/rand"
    crand "crypto/rand"
)


///////////////////////
//                  //
// Block Chain     //
//                //
///////////////////


type message_type int

const (
    INFO_MESSAGE message_type = iota
    ERROR_MESSAGE
    ATTACH_MESSAGE
)

type BlockChain struct {
    HeadSignature []byte 
    HeadBlock     *Block
}

type Part struct {
    // "Hardware" fingerprint
    Fingerprint  *big.Int

    // Part information
    PartName     string
    Manufacturer string
    SerialNumber string
    ModelNumber  string

    ProofTime    uint8
}

type Block struct {
    // Preamble
    Author            *big.Int
    Timestamp         int64
    Index             uint64
    PreviousSignature []byte
    PreviousBlock     *Block // Makes searching faster

    // Payload
    PayloadMessage    *Message

    // Apendix
    Signature         []byte
}

type Message struct {
    MessageType   message_type
    MessageString string
}

func CreatePart(partName string, manufacturer string, serialNumber string, modelNumber string) *Part {
    p := Part{}
    p.Fingerprint, _ = crand.Prime(crand.Reader, 128)
    p.PartName = partName
    p.Manufacturer = manufacturer
    p.SerialNumber = serialNumber
    p.ModelNumber = modelNumber
    return &p
}

func (p *Part) CreateBlock(payloadMessage *Message) *Block {
    b := Block{}
    b.Author = p.Fingerprint
    b.PayloadMessage = payloadMessage
    return &b
}

func (b *Block) AddBlock(previous *Block) {
    if b.Timestamp == 0 { b.Timestamp = time.Now().UnixNano() }
    
    if previous == nil {
        b.PreviousSignature = []byte{0}
        b.Index = 0
    } else {
        b.PreviousSignature = previous.Signature
        b.Index = previous.Index + 1
    }

    b.PreviousBlock = previous

    b.Signature = b.HashBlock()
}

func (b *Block) HashBlock() []byte {
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%d%d%d%x", b.Author, b.Timestamp, b.Index, b.PreviousSignature)))

    return h.Sum(nil)
}

func CompareHashes(hash1 []byte, hash2 []byte) bool {
    if hash1 == nil && hash2 == nil {
        return true
    }

    for i := 0; i < len(hash1); i++ {
        if hash1[i] != hash2[i] {
            return false
        }
    }
    return true
}

func (b *Block) StringifyBlock() string {
    s := ""
    s += fmt.Sprintf("============== Block Dump ==============\n")
    if b != nil {
        s += fmt.Sprintf(" Premble\n")
        s += fmt.Sprintf("   - Author: 0x%x\n", b.Author)
        s += fmt.Sprintf("   - Timestamp: %d\n", b.Timestamp)
        s += fmt.Sprintf("   - Index: %d\n", b.Index)
        s += fmt.Sprintf("   - Previous Block's Signature: 0x%x\n", b.PreviousSignature)
        s += fmt.Sprintf(" Payload\n")
        s += fmt.Sprintf("   - Message Payload Type: %d\n", b.PayloadMessage.MessageType)
        s += fmt.Sprintf("   - Message Payload String: %s\n", b.PayloadMessage.MessageString)
        s += fmt.Sprintf(" Appendix\n")
        s += fmt.Sprintf("   - Message Signature: 0x%x\n", b.Signature)
    } else {
        s += fmt.Sprintf(" Block in NIL\n")
    }
    s += fmt.Sprintf("\n")

    return s
}

func CreateBlockChain() *BlockChain {
    bl := BlockChain {nil, nil}

    return &bl
}

func (bl *BlockChain) AddBlock(b *Block) {
    if bl.HeadBlock == nil {

    }
    b.AddBlock(bl.HeadBlock)

    bl.HeadSignature = b.Signature
    bl.HeadBlock = b
}



///////////////////////
//                  //
// Nodes           //
//                //
///////////////////

var BlockRequestChannels []chan Block

var MAX_PRIORITY = 128

type Node struct {
    BlockRequestChannel chan Block
    BlockChain          *BlockChain
    NodePart            *Part
    Priority            int
}

func CreateNode(partName string, manufacturer string, serialNumber string, modelNumber string) *Node {
    n := Node{ }
    n.BlockChain = CreateBlockChain()
    n.Priority = rand.Intn(MAX_PRIORITY)
    n.NodePart = CreatePart(partName, manufacturer, serialNumber, modelNumber)
    n.BlockRequestChannel = make(chan Block, 16)
    BlockRequestChannels = append(BlockRequestChannels, n.BlockRequestChannel)

    return &n
}

func (n *Node) ValidateBlocks() {
    for {
        requestedBlock := <- n.BlockRequestChannel

        if n.BlockChain.HeadBlock == nil || (requestedBlock.Index == n.BlockChain.HeadBlock.Index + 1 && CompareHashes(requestedBlock.PreviousSignature, n.BlockChain.HeadBlock.Signature)) {
            n.BlockChain.AddBlock(&requestedBlock)
            for i := 0; i < len(BlockRequestChannels); i++ {
                if BlockRequestChannels[i] != n.BlockRequestChannel {
                    BlockRequestChannels[i] <- *n.BlockChain.HeadBlock
                }
            }
        }
    }
}

func (n *Node) TransmitBlock(node *Node, message *Message) { // A node can never validate it's own transactions, this improves reliablilty and 
    block := n.NodePart.CreateBlock(message)
    n.BlockChain.AddBlock(block)
    node.BlockRequestChannel <- *n.BlockChain.HeadBlock
    time.Sleep(10 * time.Millisecond)
}

func (n *Node) StringifyLedger() string {
    s := ""
    for traversalBlock := n.BlockChain.HeadBlock; traversalBlock != nil; traversalBlock = traversalBlock.PreviousBlock {
        s += traversalBlock.StringifyBlock()
    }
    return s
}


///////////////////////
//                  //
// Frontend        //
//                //
///////////////////

// Plane image was taken from here: http://dduino.blogspot.com/2012_08_01_archive.html

var STATIC_ASSETS_DIR = "/home/sckzor/Code/BlockchainInventory/static/"
var TEMPLATES_DIR = "/home/sckzor/Code/BlockchainInventory/templates/"

type TemplateData struct {
    NodesList  []*Node
    LedgerDump string
}

var GlobalTemplateData = TemplateData{  }

type neuteredFileSystem struct {
    fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
    f, err := nfs.fs.Open(path)
    if err != nil {
        return nil, err
    }

    s, err := f.Stat()
    if err != nil {
        return nil, err
    }
    if s.IsDir() {
        return nil, os.ErrNotExist
    }

    return f, nil
}

func Render(w http.ResponseWriter) {
    lp := filepath.Join(TEMPLATES_DIR, "main.html")
    
    tmpl, _ := template.ParseFiles(lp)

    tmpl.ExecuteTemplate(w, "layout", GlobalTemplateData)
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
       r.ParseForm()
        for i := 0; i < len(GlobalTemplateData.NodesList); i++ {
            if GlobalTemplateData.NodesList[i].NodePart.PartName == r.FormValue("node") {
                mt := ERROR_MESSAGE

                switch r.FormValue("severity") {
                    case "error": mt = ERROR_MESSAGE
                    case "info": mt = INFO_MESSAGE
                    case "attach": mt = ATTACH_MESSAGE 
                }

                m := Message{ mt, r.FormValue("message") }
                GlobalTemplateData.NodesList[i].TransmitBlock(GlobalTemplateData.NodesList[(i + 1) % (len(GlobalTemplateData.NodesList) - 1)], &m)
            }
        }
    }
    
    Render(w)
}


func BlocksHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        r.ParseForm()
        for i := 0; i < len(GlobalTemplateData.NodesList); i++ {
            if GlobalTemplateData.NodesList[i].NodePart.PartName == r.FormValue("node") {
                GlobalTemplateData.LedgerDump = GlobalTemplateData.NodesList[i].StringifyLedger()
            }
        }
    }
    
    Render(w)
}

func main() {
    rand.Seed(time.Now().UnixNano())

    // Create all of the nodes in the demo app
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Vertical Stabilizer",   "LKYP-U78J-6W54-MHBP", "MX540", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Rudder",                "LBMR-HLXA-GQH5-6TQ7", "TY850", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Elevator",              "PFEW-8XEJ-RZ2J-LUPA", "UL100", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Horizontal Stabilizer", "JZUA-FQVM-EH5B-E7DY", "KH300", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Ailerons",              "W4TP-7SHQ-B27V-TZGW", "KL403", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Wing",                  "YYG4-3F7B-TNXW-ALQE", "CX410", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Landing Gear",          "LSRK-TJHK-NWEP-9WR4", "VB930", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Fuselage",              "MDX6-L28Z-6GXS-FKYC", "ZD020", "Sckzor Industries"))
    GlobalTemplateData.NodesList = append(GlobalTemplateData.NodesList, CreateNode("Propellor",             "BUYZ-U64Y-O0B3-U771", "UI960", "Sckzor Industries"))

    for i := 0; i < len(GlobalTemplateData.NodesList); i++ {
        go GlobalTemplateData.NodesList[i].ValidateBlocks() // Fire up the virtual block validators
    }

    // Make all of the nodes anounce their existance to force some activity on the blockchain
    for i := 0; i < len(GlobalTemplateData.NodesList); i++ {
        m := Message{ ATTACH_MESSAGE, "The " + GlobalTemplateData.NodesList[i].NodePart.PartName + " has joined the blockchain" }
        GlobalTemplateData.NodesList[i].TransmitBlock(GlobalTemplateData.NodesList[(i + 1) % (len(GlobalTemplateData.NodesList) - 1)], &m)
        // Transmit the block to the next hop in the chain to validate it's transactions
    }

    GlobalTemplateData.LedgerDump = GlobalTemplateData.NodesList[0].StringifyLedger() // Default value for the ledger dump

    mux := http.NewServeMux()
    sfs := http.FileServer(neuteredFileSystem{http.Dir(STATIC_ASSETS_DIR)})
    
    mux.Handle("/static/", http.StripPrefix("/static/", sfs))

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { Render(w) })
    mux.HandleFunc("/message", MessageHandler)
    mux.HandleFunc("/blocks", BlocksHandler)
    
    log.Fatal(http.ListenAndServe(":8000", mux))
}

/*
    log.Println("Creating Nodes")
    n1 := CreateNode("Engine", "BUYZ-U64Y-O0B3-U771", "MX540", "Sckzor Industries")
    n2 := CreateNode("Chassis", "CV8W-JKLF-32US-18F0", "SY330", "Sckzor Industries")
    n3 := CreateNode("Transmission", "EOL2-03GH-0Y3P-SD7U", "UV871", "Sckzor Industries")
 
    log.Println("Starting Block Validation")
    go n1.ValidateBlocks()
    go n2.ValidateBlocks()
    go n3.ValidateBlocks()

    log.Println("Creating Messages To Send")
    m1 := Message{ ATTACH_MESSAGE, "The engine part has joined the blockchain" } 
    m2 := Message{ ATTACH_MESSAGE, "The transmission part has joined the blockchain" }
    m3 := Message{ ERROR_MESSAGE, "The engine is failing!" }

    log.Println("Transmitting Blocks")
    n1.TransmitBlock(n2, &m1)
    n3.TransmitBlock(n1, &m2)
    n1.TransmitBlock(n3, &m3)

    time.Sleep(time.Second / 2)

    log.Println("============ Printing Ledger For Node 1 ============")
    n1.PrintLedger()

    log.Println("============ Printing Ledger For Node 2 ============")
    n2.PrintLedger()
    
    log.Println("============ Printing Ledger For Node 3 ============")
    n3.PrintLedger()

    for {  }
*/


/*
    bl := CreateBlockChain()

    p1 := CreatePart("Engine", "0000-0000-0000-0000", "MX540", "Sckzor Industries")
    m1 := Message{ ATTACH_MESSAGE, "The engine part has joined the blockchain" }
    b1 := p1.CreateBlock(&m1)
    bl.AddBlock(b1)

    p2 := CreatePart("Steering Wheel", "1000-0000-0000-0000", "SY330", "Sckzor Industries")
    m2 := Message{ ATTACH_MESSAGE, "The steering wheel part has joined the blockchain" }
    b2 := p2.CreateBlock(&m2)
    bl.AddBlock(b2)

    m3 := Message{ ERROR_MESSAGE, "The engine is failing!" }
    b3 := p1.CreateBlock(&m3)
    bl.AddBlock(b3)

    bl.HeadBlock.PrintBlock()
    bl.HeadBlock.PreviousBlock.PrintBlock()
    bl.HeadBlock.PreviousBlock.PreviousBlock.PrintBlock()
*/
