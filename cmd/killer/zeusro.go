package killer

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/config"
)

// Coin live or dead ?
func (z *Zeusro) Coin() bool {
	//prevent same result
	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(2)
	return v == 1
}

// Zeusro Kuiper
type Zeusro struct {
	dryRun bool
	config *config.ProjectConfig
}

func NewZeusro(config *config.ProjectConfig, dryRun bool) *Zeusro {
	z := Zeusro{
		dryRun: dryRun,
		config: config,
	}
	return &z
}

func (z *Zeusro) Run() {
	if z.dryRun {
		fmt.Println("Leonard: I am sorry,I have to go.")
		time.Sleep(time.Second)
		fmt.Println("Leonard: I don't believe this.")
		time.Sleep(time.Second * 2)
		z.callSheldon()
		return
	}
	coin := z.Coin()
	dead := !coin
	if dead {
		fmt.Println("Zeusro: Goodbye.")
		z.callMyWife()
	}
	live := coin
	if live {
		z.callAryaStark()
	}
}

func (z *Zeusro) callAryaStark() {
	coin := z.Coin()
	dead := !coin
	if dead {
		fmt.Println("Arya: Valar Morghulis")
		z.callThanos()
		return
	}
	live := coin
	if live {
		fmt.Println("Arya: Valar Dohaeris")
		return
	}
}

func (z *Zeusro) callSheldon() {
	fmt.Print("Sheldon: A")
	for i := 0; i < 2049; i++ {
		time.Sleep(time.Millisecond * 2)
		fmt.Print("a")
	}
	fmt.Println("!")
	time.Sleep(time.Second * 3)
	fmt.Println("Sheldon: BAZINGA PUNK!!! NOW WE'RE EVEN.")
}

func (z *Zeusro) callMyWife() {
	//TODO
}

func (z *Zeusro) callThanos() {
	//TODO
}
