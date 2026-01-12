package killer

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/config"
	"github.com/rs/zerolog/log"
)

// Coin returns true with 50% probability
func (z *Zeusro) Coin() bool {
	// Use rand.New() instead of deprecated rand.Seed() (Go 1.20+)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	v := r.Intn(2)
	return v == 1
}

// Zeusro Kuiper - The unpredictable killer
type Zeusro struct {
	dryRun    bool
	namespace string
	config    *config.ProjectConfig
}

func NewZeusro(config *config.ProjectConfig, namespace string, dryRun bool) *Zeusro {
	z := Zeusro{
		dryRun:    dryRun,
		namespace: namespace,
		config:    config,
	}
	return &z
}

// DryRun sets the dryRun flag
func (z *Zeusro) DryRun() *Zeusro {
	z.dryRun = true
	return z
}

// Run executes the Zeusro command:
// - 50% probability: Valar Dohaeris (nothing happens)
// - 50% probability: Thanos mode (randomly delete 50% of pods)
func (z *Zeusro) Run() error {
	if z.dryRun {
		log.Info().Msg("Zeusro: [DRY RUN] The coin will be flipped...")
		coin := z.Coin()
		if coin {
			log.Info().Msg("Arya: Valar Dohaeris (Nothing happens)")
		} else {
			log.Warn().Msg("Thanos: I am inevitable. [DRY RUN] Would delete 50% of pods")
		}
		return nil
	}

	coin := z.Coin()
	if coin {
		// Valar Dohaeris - nothing happens
		fmt.Println("Arya: Valar Dohaeris")
		log.Info().Msg("Zeusro: All shall serve. Nothing happens.")
		return nil
	}

	// Thanos mode - delete 50% of pods
	fmt.Println("Thanos: I am inevitable.")
	time.Sleep(time.Second)
	fmt.Println("Thanos: *snaps fingers*")
	time.Sleep(time.Second)
	fmt.Println("Thanos: The universe will be balanced.")
	log.Warn().Msg("Zeusro: Thanos mode activated - deleting 50% of pods randomly")

	// Create PodKiller and kill half of the pods
	podKiller, err := NewPodKiller(z.namespace)
	if err != nil {
		return fmt.Errorf("failed to create PodKiller: %w", err)
	}

	// Set half mode to delete 50% of pods
	podKiller.SetHalf()
	// BlackHand is needed to trigger KillHalfPods
	podKiller.BlackHand()

	return podKiller.Kill()
}

// Legacy methods kept for backward compatibility
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

func (z *Zeusro) callThanos() {
	fmt.Println("Thanos: I am inevitable.")
	time.Sleep(time.Second * 2)
	fmt.Println("Thanos: *snaps fingers*")
	time.Sleep(time.Second)
	fmt.Println("Thanos: The universe will be balanced.")
}
