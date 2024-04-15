package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Define os elementos do jogo
type Elemento struct {
    simbolo rune
    cor termbox.Attribute
    corFundo termbox.Attribute
    tangivel bool
}

// Personagem controlado pelo jogador
var personagem = Elemento{
    simbolo: '☺',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

// Parede
var parede = Elemento{
    simbolo: '▣',
    cor: termbox.ColorDefault,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

// Barrreira
var barreira = Elemento{
    simbolo: '#',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

// Vegetação
var vegetacao = Elemento{
    simbolo: '♣',
    cor: termbox.ColorGreen,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Elemento vazio
var vazio = Elemento{
    simbolo: ' ',
    cor: termbox.ColorDefault,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

// Elemento para representar áreas não reveladas (efeito de neblina)
var neblina = Elemento{
    simbolo: '.',
    cor: termbox.ColorDefault,
    corFundo: termbox.ColorYellow,
    tangivel: false,
}


var fogo = Elemento{
    simbolo: '♨',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

var fogoApagado = Elemento{
    simbolo: '⧇',
    cor: termbox.ColorLightGray,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var bastaoAgua = Elemento{
    simbolo: '☿',
    cor: termbox.ColorBlue,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var agua = Elemento{
    simbolo: '◯',
    cor: termbox.ColorBlue,
    corFundo: termbox.ColorDefault,
    tangivel: false,
}

var inimigo = Elemento{
    simbolo: '☠',
    cor: termbox.ColorRed,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

var npc = Elemento{
    simbolo: '♗',
    cor: termbox.ColorLightYellow,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}

var final = Elemento{
    simbolo: '⛤',
    cor: termbox.ColorYellow,
    corFundo: termbox.ColorDefault,
    tangivel: true,
}



var mapa [][]Elemento
var posX, posY int
var bastaoX, bastaoY int
var aguaX, aguaY int
var posicoesInimigos [][]int
var temBastaoAgua bool = false
var temDialogo bool = true
var ultimoElementoSobPersonagem = vazio
var statusMsg string
var vidas int = 3
var ultimaDirecao string
var efeitoNeblina = false
var revelado [][]bool
var raioVisao int = 3


var mapaMutex sync.Mutex


func mudaFogo() {
    for {
        for y, linha := range mapa {
            for x, elem := range linha {
                if elem == fogo {
                    mapa[y][x] = fogoApagado
                    } else if elem == fogoApagado {
                        mapa[y][x] = fogo
                    }
                }
            }
            mapaMutex.Lock()
            desenhaTudo()
            mapaMutex.Unlock()
            time.Sleep(2 * time.Second)
        }
}


func main() {

    

    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    carregarMapa("mapa.txt")
    if efeitoNeblina {
        revelarArea()
    }
    mapaMutex.Lock()
    desenhaTudo()
    mapaMutex.Unlock()
    go mudaFogo()
    moverInimigos()

    for {
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
            if ev.Key == termbox.KeyEsc {
                return // Sair do programa
            }
            if ev.Ch == 'e' {
                interagir()
            } else {
                mover(ev.Ch)
                if efeitoNeblina {
                    revelarArea()
                }
            }
            mapaMutex.Lock()
            desenhaTudo()
            mapaMutex.Unlock()
        }
    }
}

func carregarMapa(nomeArquivo string) {
    arquivo, err := os.Open(nomeArquivo)
    if err != nil {
        panic(err)
    }
    defer arquivo.Close()

    scanner := bufio.NewScanner(arquivo)
    y := 0
    for scanner.Scan() {
        linhaTexto := scanner.Text()
        var linhaElementos []Elemento
        var linhaRevelada []bool
        for x, char := range linhaTexto {
            elementoAtual := vazio
            switch char {
            case parede.simbolo:
                elementoAtual = parede
            case fogo.simbolo:
                elementoAtual = fogo
            case fogoApagado.simbolo:
                elementoAtual = fogoApagado
            case agua.simbolo:
                aguaX, aguaY = x, y
                elementoAtual = vazio
            case barreira.simbolo:
                elementoAtual = barreira
            case npc.simbolo:
                elementoAtual = npc
            case inimigo.simbolo:
                posicoesInimigos = append(posicoesInimigos, []int{x, y})
                elementoAtual = vazio
            case vegetacao.simbolo:
                elementoAtual = vegetacao
            case final.simbolo:
                elementoAtual = final
            case bastaoAgua.simbolo:
                elementoAtual = bastaoAgua
                bastaoX, bastaoY = x, y
            case personagem.simbolo:
                // Atualiza a posição inicial do personagem
                posX, posY = x, y
                elementoAtual = vazio
            }
            linhaElementos = append(linhaElementos, elementoAtual)
            linhaRevelada = append(linhaRevelada, false)
        }
        mapa = append(mapa, linhaElementos)
        revelado = append(revelado, linhaRevelada)
        y++
    }
    if err := scanner.Err(); err != nil {
        panic(err)
    }
}

func desenhaTudo() {
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    for y, linha := range mapa {
        for x, elem := range linha {
            if efeitoNeblina == false || revelado[y][x] {
                termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
            } else {
                termbox.SetCell(x, y, neblina.simbolo, neblina.cor, neblina.corFundo)
            }
        }
    }

    desenhaBarraDeStatus()

    termbox.Flush()
}

func desenhaBarraDeStatus() {
    for i, c := range fmt.Sprintf("Vidas: %d", vidas) {
        termbox.SetCell(i, len(mapa), c, termbox.ColorWhite, termbox.ColorDefault)
    }
    for i, c := range statusMsg {
        termbox.SetCell(i, len(mapa)+1, c, termbox.ColorWhite, termbox.ColorDefault)
    }
    msg := "Use WASD para mover e E para interagir. ESC para sair."
    for i, c := range msg {
        termbox.SetCell(i, len(mapa)+3, c, termbox.ColorWhite, termbox.ColorDefault)
    }
}

func revelarArea() {
    minX := max(0, posX-raioVisao)
    maxX := min(len(mapa[0])-1, posX+raioVisao)
    minY := max(0, posY-raioVisao/2)
    maxY := min(len(mapa)-1, posY+raioVisao/2)

    for y := minY; y <= maxY; y++ {
        for x := minX; x <= maxX; x++ {
            // Revela as células dentro do quadrado de visão
            revelado[y][x] = true
        }
    }
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func mover(comando rune) {
    dx, dy := 0, 0
    switch comando {
    case 'w':
        dy = -1
        ultimaDirecao = "cima"
    case 'a':
        dx = -1
        ultimaDirecao = "esquerda"
    case 's':
        dy = 1
        ultimaDirecao = "baixo"
    case 'd':
        dx = 1
        ultimaDirecao = "direita"
    }
    novaPosX, novaPosY := posX+dx, posY+dy
    if mapa[novaPosY][novaPosX] == fogo {
        vidas--
        statusMsg = "Você foi queimado!"
        if vidas == 0 {
            statusMsg = "Você morreu!"
            time.Sleep(2 * time.Second)
            os.Exit(0)
        }
     }
    if mapa[novaPosY][novaPosX] == final {
        statusMsg = "Você Ganhou!"
            time.Sleep(2 * time.Second)
            os.Exit(0)
     }
    if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
        mapa[novaPosY][novaPosX].tangivel == false {
        mapa[posY][posX] = ultimoElementoSobPersonagem // Restaura o elemento anterior
        ultimoElementoSobPersonagem = mapa[novaPosY][novaPosX] // Atualiza o elemento sob o personagem
        posX, posY = novaPosX, novaPosY // Move o personagem
        mapa[posY][posX] = personagem // Coloca o personagem na nova posição
    }
}

func moverInimigos() {
    for _, posicao := range posicoesInimigos {
        go moverInimigo(posicao[0], posicao[1])
    }
}



func moverInimigo(xInicial, yInicial int) {
    var inimigoX, inimigoY int = xInicial, yInicial
    var velocidade int = rand.Intn(501) + 500
    // o inimigo se move aleatoriamente em uma direção
    // até encontrar um obstáculo ou fogo e muda de direção
    for {
        dx, dy := 0, 0
        switch rand.Intn(4) {
        case 0:
            dy = -1
        case 1:
            dx = -1
        case 2:
            dy = 1
        case 3:
            dx = 1
        }
        novaPosX, novaPosY := inimigoX+dx, inimigoY+dy
        if mapa[novaPosY][novaPosX] == personagem {
            vidas--
            statusMsg = "Você foi atacado!"
            if vidas == 0 {
                statusMsg = "Você morreu!"
                time.Sleep(2 * time.Second)
                os.Exit(0)
            }
        }

        if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
            mapa[novaPosY][novaPosX].tangivel == false && mapa[novaPosY][novaPosX] != fogoApagado{
            mapa[inimigoY][inimigoX] = vazio
            inimigoX, inimigoY = novaPosX, novaPosY
            mapa[inimigoY][inimigoX] = inimigo
        
        }
        mapaMutex.Lock()
        desenhaTudo()
        mapaMutex.Unlock()
        time.Sleep(time.Duration(velocidade) * time.Millisecond)
    }
}



func pegaBastaoAgua() {
                statusMsg = "Pegou cajado de agua, aperte E para atirar."
                temBastaoAgua = true
}

func dispararAgua() {
    if !temBastaoAgua || (aguaX != 0 && aguaY != 0) {
        return // Saia se o jogador não tiver o bastão de água ou se já houver água no tabuleiro
    }

    
    dx, dy := 0, 0
    switch ultimaDirecao {
    case "cima":
        dy = -1
    case "esquerda":
        dx = -1
    case "baixo":
        dy = 1
    case "direita":
        dx = 1
    }
    
    aguaX, aguaY = posX, posY
    if mapa[aguaY+dy][aguaX+dx] == vazio {
    mapa[aguaY+dy][aguaX+dx] = agua // Atribui a água à posição inicial
    }


    for {

        novaPosX, novaPosY := aguaX+dx, aguaY+dy
        if(mapa[novaPosY][novaPosX] == fogo || mapa[novaPosY][novaPosX] == fogoApagado) {
            mapa[novaPosY][novaPosX] = vazio
            mapa[aguaY][aguaX] = vazio
            aguaX, aguaY = 0, 0
            statusMsg = "Apagou o fogo"
            return
        }
        if novaPosY >= 0 && novaPosY < len(mapa) && novaPosX >= 0 && novaPosX < len(mapa[novaPosY]) &&
            mapa[novaPosY][novaPosX].tangivel == false {
            // Verifica se a posição anterior está vazia antes de atribuir água à nova posição
            if mapa[aguaY][aguaX] != personagem {
                mapa[aguaY][aguaX] = vazio
            }
            aguaX, aguaY = novaPosX, novaPosY // Move o tiro
            mapa[aguaY][aguaX] = agua         // Coloca a água na nova posição
        } else {
            // Se a nova posição não for válida, remove a água do mapa e redefine sua posição para 0,0
            mapa[aguaY][aguaX] = vazio
            aguaX, aguaY = 0, 0
            return
        }

        mapaMutex.Lock()
        desenhaTudo()
        mapaMutex.Unlock()
        time.Sleep(time.Millisecond * 50) // Ajuste do tempo de espera
    }
}


func containsElement(slice []Elemento, val Elemento) bool {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}
        
func interagir() {
    var elementosAoRedor []Elemento
    var bastaoX, bastaoY int

    // verifica se há elementos em uma área de 3x3 ao redor do personagem
    for y := posY - 1; y <= posY + 1; y++ {
        for x := posX - 1; x <= posX + 1; x++ {
            if y >= 0 && y < len(mapa) && x >= 0 && x < len(mapa[y]) {
                if mapa[y][x] != vazio {
                    if mapa[y][x] == bastaoAgua {
                        bastaoX, bastaoY = x, y
                    }
                elementosAoRedor = append(elementosAoRedor, mapa[y][x])
                }
            }
        }
    }

    if containsElement(elementosAoRedor, npc) && temDialogo{
        temDialogo = false
        statusMsg = "O NPC diz: 'Interaja com o cajado de água para disparar água no fogo.'"
    } else {
        for _, item := range elementosAoRedor {
            if item.simbolo == '☿' {
            statusMsg = "Pegou cajado de agua, aperte E para atirar."
            temBastaoAgua = true
            mapa[bastaoY][bastaoX] = vazio
        }
    }
}
    if temBastaoAgua {
        go dispararAgua()
    }
}