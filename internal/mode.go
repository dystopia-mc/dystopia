package internal

var FlyingMode = flyingMode{}

type flyingMode struct{}

func (flyingMode) AllowsEditing() bool      { return false }
func (flyingMode) AllowsTakingDamage() bool { return false }
func (flyingMode) CreativeInventory() bool  { return false }
func (flyingMode) HasCollision() bool       { return true }
func (flyingMode) AllowsInteraction() bool  { return false }
func (flyingMode) Visible() bool            { return true }
func (flyingMode) AllowsFlying() bool       { return true }
