package main

import (
	"time"
	"sync"
)

// Effects that can be applied.
type StatusEffectType string
const ( //Building it this way avoids using string values on maps but allows the json to bear the string value.
	Dazed StatusEffectType = "dazed"
	Blind StatusEffectType = "blind"
	Poison StatusEffectType = "poison"
	Bleed StatusEffectType = "bleed"
)

// Information on single status effect.
type StackInfo struct {
	Count int64 `json:"count"`
	Max int64 `json:"max"`
	Chance float64 `json:"chance"`
}

// Status effects data structure.
type StatusEffect struct {
	Type StatusEffectType `json:"type"`
	Modifier float64 `json:"modifier"` //Percent based modifier as a float.  Ex: -50% -> 0.5.
	Duration int64 `json:"duration"` //How long does the effect last for in seconds.
	Interval int64 `json:"interval"` //How many seconds of the duraction until the modifier is applied.
	Stack StackInfo `json:"stack"` //Modifier multiplier.
	ExpiresAt int64 `json:"expires_at"` //Timestamp of when the effect will fall off.
}

// Registry to hold all of the definitions.  Using a mutex here since the data could be live-ops driven meaning it could change after nakama init.
// **NOTE: If the plan is to not update this information after nakama init then this paradigm can be change to a simple read-only map instead.
var StatusEffectsRegistry = struct {
	sync.RWMutex //Read/write mutex to help with concurrent access allowing mulitple readers or a single writer.
	StatusEffects map[StatusEffectType]StatusEffect
}{
	StatusEffects: make(map[StatusEffectType]StatusEffect),
}

// This function will initialize the Attack Registry.
// @JWK TODO: Move this information into storage instead of hard coded value to place it under the purvue of live-ops.
func InitStatusEffectsRegistry() {
	StatusEffectsRegistry.Lock() //Call lock on the mutex in preparation for writing.
	defer StatusEffectsRegistry.Unlock() //Don't forget to release the mutex lock.
	StatusEffectsRegistry.StatusEffects[Dazed] = StatusEffect{
		Type: Dazed,
		Modifier: -0.5, //Modifier that will be used in the game logic.
		Duration: 30, //Seconds
		Interval: 0, //Seconds
		Stack: StackInfo{},
		ExpiresAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Blind] = StatusEffect{
		Type: Blind,
		Modifier: -0.95, //Modifier that will be used in the game logic.
		Duration: 10, //Seconds
		Interval: 0, //Seconds
		Stack: StackInfo{},
		ExpiresAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Poison] = StatusEffect{
		Type: Poison,
		Modifier: -5, //Modifier that will be used in the game logic.
		Duration: 30, //Seconds
		Interval: 0, //Seconds
		Stack: StackInfo{},
		ExpiresAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Bleed] = StatusEffect{
		Type: Bleed,
		Modifier: -2, //Modifier that will be used in the game logic.
		Duration: 60, //Seconds
		Interval: 0, //Seconds
		Stack: StackInfo{
			Count: 0, //This gets set if it is applied.
			Max: 3, //Maximum possible stacks.
			Chance: 0.6, //Chance to apply.
		},
		ExpiresAt: 0, //This gets set when it is applied.
	}
}

// This function adds a status effect to a target.  
func (p *Player) AddStatusEffect(effect StatusEffectType, modifier float64, duration int64) {
	timestamp := time.Now().Unix()
	//Append the status effect to the players effects.
	p.StatusEffects = append(p.StatusEffects, StatusEffect{
		Type: effect,
		Modifier: modifier,
		Duration: duration,
		ExpiresAt: timestamp + (duration * 1000), //Convert to milliseconds.
	})
	p.UpdatedAt = timestamp
}

// This function processes status effects, removes expired status effects, decrements duration to help the client anticipate fall off.
func (p *Player) TickStatusEffect() {
	timestamp := time.Now().Unix()
	//Check if there are any effects to process.
	if len(p.StatusEffects) > 0 {
		var processedEffects []StatusEffect
		//Loop over the status effects to process them.
		for _, effect := range p.StatusEffects {
			//If there effect hasn't expired then process it, otherwise it falls off.
			if effect.ExpiresAt > timestamp {
				//Update duraction with time delta.
				effect.Duration = (effect.ExpiresAt - timestamp) / 1000 //Keep in milliseconds.
				processedEffects = append(processedEffects, effect)
			}
		}
		p.StatusEffects = processedEffects
		p.UpdatedAt = timestamp
	}
}
