package main

import (
	"time"
	"sync"
	"math"

	"github.com/heroiclabs/nakama-common/runtime"
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
	UpdatedAt int64 `json:"updated_at"` //Timestamp of when the effects were last processed.
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
		UpdatedAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Blind] = StatusEffect{
		Type: Blind,
		Modifier: -0.95, //Modifier that will be used in the game logic.
		Duration: 10, //Seconds
		Interval: 0, //Seconds
		Stack: StackInfo{},
		ExpiresAt: 0, //This gets set when it is applied.
		UpdatedAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Poison] = StatusEffect{
		Type: Poison,
		Modifier: -5, //Modifier that will be used in the game logic.
		Duration: 30, //Seconds
		Interval: 3, //Seconds
		Stack: StackInfo{},
		ExpiresAt: 0, //This gets set when it is applied.
		UpdatedAt: 0, //This gets set when it is applied.
	}
	StatusEffectsRegistry.StatusEffects[Bleed] = StatusEffect{
		Type: Bleed,
		Modifier: -2, //Modifier that will be used in the game logic.
		Duration: 60, //Seconds
		Interval: 5, //Seconds
		Stack: StackInfo{
			Count: 0, //This gets set if it is applied.
			Max: 3, //Maximum possible stacks.
			Chance: 0.6, //Chance to apply.
		},
		ExpiresAt: 0, //This gets set when it is applied.
		UpdatedAt: 0, //This gets set when it is applied.
	}
}

// This function add status effects.  
func AddStatusEffect(logger runtime.Logger, effectType StatusEffectType, ep EntityProcessor) {
	//@JWK TODO: Handle stacking effects.
	//@JWK TODO: If it's already an existing effect but doesn't stack, refresh duration and ExpiresAt..
	timestamp := time.Now().Unix()
	logger.Debug("Adding status effect at: %v", timestamp)
	StatusEffectsRegistry.RLock() //Read lock
	defer StatusEffectsRegistry.RUnlock() //Release lock
	effect := StatusEffectsRegistry.StatusEffects[effectType]
	if effect.Type == "" {
		logger.Error("Status effect type NOT found: %s", effectType)
		return
	}
	duration := (effect.Duration * 10) //@JWK TODO: Remove multiplier.
	effect.Duration = duration  //@JWK TODO: Remove me when multiplier is gone.
	effect.ExpiresAt = timestamp + duration
	effect.UpdatedAt = timestamp
	//Append and set.
	logger.Debug("Added status effect: %+v", effect)
	statusEffects := ep.GetStatusEffects()
	statusEffects = append(statusEffects, &effect)
	ep.SetStatusEffects(statusEffects)
}

// This function removes expired status effects and decrements duration to help the client anticipate fall off.
func TickStatusEffect(logger runtime.Logger, ep EntityProcessor) {
	timestamp := time.Now().Unix()
	//Check if there are any effects to process.
	statusEffects := ep.GetStatusEffects()
	if len(statusEffects) > 0 {
		logger.Debug("Tick status effects for source")
		var processedEffects []*StatusEffect
		//Loop over the status effects to process them.
		for _, effect := range statusEffects {
			//Process effects that deal damage over time.
			if effect.Type == Bleed || effect.Type == Poison {
				logger.Debug("Tick status effects, processing: %+v", effect)
				//@JWK TODO: Check for stacks.
				health := ep.GetHealth()
				//Get delta .
				delta := (timestamp - effect.UpdatedAt) //@JWK TODO: Find the delta between the update_at and timestamp and compare that value to the duration for gauging if effect should be "refreshed".
				logger.Debug("Tick delta(%d) = update(%d) / timestamp(%d)", delta, effect.UpdatedAt, timestamp)
				if delta >= 0 { //Positive means future date.
					//Calculate the intervals of period damage.
					if effect.Interval == 0 {
						logger.Error("Effect interval 0, can't divide by 0: %+v", effect)
						continue
					}
					intervals := delta / effect.Interval //Go rounds down so don't have to floor.
					logger.Debug("Tick intervals(%d) = delta(%d) / interval(%d)", intervals, delta, effect.Interval)
					//Caclulate the total damage.
					damage := intervals * int64(effect.Modifier)
					logger.Debug("Tick damage(%d) = intervals(%d) * modifier(%f)", damage, intervals, effect.Modifier)
					//Calculate max damage.
					mDamage := (effect.Duration / effect.Interval) * int64(effect.Modifier)
					logger.Debug("Tick mDamage(%d) = (duration(%d) / interval(%d)) * modifier(%f)", mDamage, effect.Duration, effect.Interval, effect.Modifier)
					if math.Abs(float64(damage)) > math.Abs(float64(mDamage)) {
						damage = mDamage //Max damage that would be applied.
					}
					//Impact health, negeative numbers will decrement.
					health += int(damage)
					logger.Debug("Tick health H:%d - D:%d", health, damage)
					ep.SetHealth(health)
				}
			}
			//If the effect hasn't expired then updated it, otherwise it falls off.
			delta := (effect.ExpiresAt - timestamp)
			if delta > 0 { //Positive means future date.
				logger.Debug("Tick status effects, processing: %+v", effect)
				//Update duraction with time delta.
				effect.Duration = delta
				effect.UpdatedAt = timestamp
				processedEffects = append(processedEffects, effect) //Keep the ones not expired.
			}
		}
		ep.SetStatusEffects(processedEffects)
	}
}

// Interface function to get status effects.
func (p *Player) GetStatusEffects() []*StatusEffect {
	return p.StatusEffects
}

// Interface function to set status effects.
func (p *Player) SetStatusEffects(statusEffects []*StatusEffect) {
	p.StatusEffects = statusEffects
}

// Interface function to get status effects.
func (e *Enemy) GetStatusEffects() []*StatusEffect {
	return e.StatusEffects
}

// Interface function to set status effects.
func (e *Enemy) SetStatusEffects(statusEffects []*StatusEffect) {
	e.StatusEffects = statusEffects
}

