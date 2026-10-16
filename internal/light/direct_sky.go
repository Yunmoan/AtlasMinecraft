package light

import "fmt"

// DirectSkylight reports whether every state on a top-to-bottom column path
// has unobstructed level-15 skylight. It uses the same original opacity and
// directional face occlusion as relight. A blocked path needs propagation to
// determine the actual value, since light may still arrive from its sides.
func DirectSkylight(topDown []uint16) (bool, error) {
	rules, err := loadRules()
	if err != nil {
		return false, err
	}
	previous := rules.states[0]
	for _, id := range topDown {
		if int(id) >= len(rules.states) {
			return false, fmt.Errorf("unknown light block state %d", id)
		}
		current := rules.states[id]
		if current.Opacity > 0 || rules.blocked[previous.Faces[0]][current.Faces[1]] {
			return false, nil
		}
		previous = current
	}
	return true, nil
}
