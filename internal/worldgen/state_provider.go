package worldgen

import (
	"atlas/internal/blockstate"
	"encoding/json"
	"fmt"
	"math"
)

type featureStateProvider func(*WorldgenRandom, FeaturePosition) SurfaceBlockState

func decodeFeatureStateProvider(raw json.RawMessage, validate func(SurfaceBlockState) error) (featureStateProvider, error) {
	var config struct {
		Type    string            `json:"type"`
		State   SurfaceBlockState `json:"state"`
		Entries []struct {
			Data   SurfaceBlockState `json:"data"`
			Weight int32             `json:"weight"`
		} `json:"entries"`
		Seed  int64 `json:"seed"`
		Noise struct {
			FirstOctave int       `json:"firstOctave"`
			Amplitudes  []float64 `json:"amplitudes"`
		} `json:"noise"`
		Scale      float32             `json:"scale"`
		States     []SurfaceBlockState `json:"states"`
		Threshold  float32             `json:"threshold"`
		Chance     float32             `json:"high_chance"`
		Default    SurfaceBlockState   `json:"default_state"`
		LowStates  []SurfaceBlockState `json:"low_states"`
		HighStates []SurfaceBlockState `json:"high_states"`
		Variety    []int32             `json:"variety"`
		SlowNoise  struct {
			FirstOctave int       `json:"firstOctave"`
			Amplitudes  []float64 `json:"amplitudes"`
		} `json:"slow_noise"`
		SlowScale float32         `json:"slow_scale"`
		Property  string          `json:"property"`
		Source    json.RawMessage `json:"source"`
		Values    json.RawMessage `json:"values"`
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}
	checkState := func(state SurfaceBlockState) error {
		if _, err := blockstate.BlockStateID(state.Name, state.Properties); err != nil {
			return err
		}
		return validate(state)
	}
	switch config.Type {
	case "minecraft:randomized_int_state_provider":
		if config.Property != "age" {
			return nil, fmt.Errorf("unsupported randomized state property %s", config.Property)
		}
		values, err := decodeFeatureIntProvider(config.Values, 0, 25)
		if err != nil {
			return nil, err
		}
		source, err := decodeFeatureStateProvider(config.Source, func(s SurfaceBlockState) error {
			if _, exists := s.Properties[config.Property]; !exists {
				return fmt.Errorf("missing randomized state property")
			}
			return validate(s)
		})
		if err != nil {
			return nil, err
		}
		return func(r *WorldgenRandom, p FeaturePosition) SurfaceBlockState {
			s := source(r, p)
			return stateWithProperty(s, config.Property, fmt.Sprint(values(r)))
		}, nil
	case "minecraft:simple_state_provider":
		if err := checkState(config.State); err != nil {
			return nil, err
		}
		return func(*WorldgenRandom, FeaturePosition) SurfaceBlockState { return config.State }, nil
	case "minecraft:weighted_state_provider":
		var total int32
		for _, entry := range config.Entries {
			if err := checkState(entry.Data); err != nil {
				return nil, err
			}
			if entry.Weight < 0 || int64(total)+int64(entry.Weight) > math.MaxInt32 {
				return nil, fmt.Errorf("invalid state provider weight")
			}
			total += entry.Weight
		}
		if total == 0 {
			return nil, fmt.Errorf("empty weighted state provider")
		}
		return func(random *WorldgenRandom, _ FeaturePosition) SurfaceBlockState {
			choice := random.NextIntBound(total)
			for _, entry := range config.Entries {
				choice -= entry.Weight
				if choice < 0 {
					return entry.Data
				}
			}
			panic("worldgen: invalid weighted state provider")
		}, nil
	case "minecraft:noise_provider", "minecraft:noise_threshold_provider", "minecraft:dual_noise_provider":
		if config.Scale <= 0 || math.IsInf(float64(config.Scale), 0) {
			return nil, fmt.Errorf("invalid noise provider scale")
		}
		noise, err := newProviderNormalNoise(config.Seed, config.Noise.FirstOctave, config.Noise.Amplitudes)
		if err != nil {
			return nil, err
		}
		sample := func(p FeaturePosition) float64 {
			return noise.Noise(float64(p.X)*float64(config.Scale), float64(p.Y)*float64(config.Scale), float64(p.Z)*float64(config.Scale))
		}
		choose := func(states []SurfaceBlockState, value float64) SurfaceBlockState {
			index := int(max(0, min(0.9999, (1+value)/2)) * float64(len(states)))
			return states[index]
		}
		if config.Type == "minecraft:noise_threshold_provider" {
			if len(config.LowStates) == 0 || len(config.HighStates) == 0 || config.Chance < 0 || config.Chance > 1 || config.Threshold < -1 || config.Threshold > 1 {
				return nil, fmt.Errorf("invalid noise threshold provider")
			}
			for _, state := range append(append([]SurfaceBlockState{config.Default}, config.LowStates...), config.HighStates...) {
				if err := checkState(state); err != nil {
					return nil, err
				}
			}
			return func(random *WorldgenRandom, p FeaturePosition) SurfaceBlockState {
				if sample(p) < float64(config.Threshold) {
					return config.LowStates[random.NextIntBound(int32(len(config.LowStates)))]
				}
				if random.NextFloat() < config.Chance {
					return config.HighStates[random.NextIntBound(int32(len(config.HighStates)))]
				}
				return config.Default
			}, nil
		}
		if len(config.States) == 0 {
			return nil, fmt.Errorf("empty noise state provider")
		}
		for _, state := range config.States {
			if err := checkState(state); err != nil {
				return nil, err
			}
		}
		if config.Type == "minecraft:noise_provider" {
			return func(_ *WorldgenRandom, p FeaturePosition) SurfaceBlockState { return choose(config.States, sample(p)) }, nil
		}
		if len(config.Variety) != 2 || config.Variety[0] < 1 || config.Variety[1] > 64 || config.Variety[0] > config.Variety[1] || config.SlowScale <= 0 {
			return nil, fmt.Errorf("invalid dual noise provider")
		}
		slow, err := newProviderNormalNoise(config.Seed, config.SlowNoise.FirstOctave, config.SlowNoise.Amplitudes)
		if err != nil {
			return nil, err
		}
		// DualNoiseProvider deliberately multiplies slow coordinates as floats.
		slowSample := func(p FeaturePosition) float64 {
			return slow.Noise(float64(float32(p.X)*config.SlowScale), float64(float32(p.Y)*config.SlowScale), float64(float32(p.Z)*config.SlowScale))
		}
		return func(_ *WorldgenRandom, p FeaturePosition) SurfaceBlockState {
			value := slowSample(p)
			count := int(float64(config.Variety[0]) + max(0, min(1, (value+1)/2))*float64(config.Variety[1]+1-config.Variety[0]))
			states := make([]SurfaceBlockState, count)
			for i := range states {
				states[i] = choose(config.States, slowSample(treeOffset(p, int32(i)*54545, 0, int32(i)*34234)))
			}
			return choose(states, sample(p))
		}, nil
	default:
		return nil, fmt.Errorf("unsupported block state provider %s", config.Type)
	}
}
