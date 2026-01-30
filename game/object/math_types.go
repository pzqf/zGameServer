package object

import (
	"math"
)

// Vector3 三维向量
type Vector3 struct {
	X, Y, Z float32
}

// NewVector3 创建新的三维向量
func NewVector3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Add 向量加法
func (v Vector3) Add(other Vector3) Vector3 {
	return Vector3{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// Sub 向量减法
func (v Vector3) Sub(other Vector3) Vector3 {
	return Vector3{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Mul 向量乘法
func (v Vector3) Mul(scalar float32) Vector3 {
	return Vector3{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// Div 向量除法
func (v Vector3) Div(scalar float32) Vector3 {
	return Vector3{
		X: v.X / scalar,
		Y: v.Y / scalar,
		Z: v.Z / scalar,
	}
}

// Length 计算向量长度
func (v Vector3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// Normalize 归一化向量
func (v Vector3) Normalize() Vector3 {
	length := v.Length()
	if length == 0 {
		return Vector3{0, 0, 0}
	}
	return v.Div(length)
}

// Distance 计算两个向量之间的距离
func (v Vector3) Distance(other Vector3) float32 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// Lerp 线性插值
func (v Vector3) Lerp(other Vector3, t float32) Vector3 {
	return Vector3{
		X: v.X + (other.X-v.X)*t,
		Y: v.Y + (other.Y-v.Y)*t,
		Z: v.Z + (other.Z-v.Z)*t,
	}
}

// Quaternion 四元数
type Quaternion struct {
	X, Y, Z, W float32
}

// NewQuaternion 创建新的四元数
func NewQuaternion(x, y, z, w float32) Quaternion {
	return Quaternion{X: x, Y: y, Z: z, W: w}
}

// EulerToQuaternion 欧拉角转四元数
func EulerToQuaternion(roll, pitch, yaw float32) Quaternion {
	cr := float32(math.Cos(float64(roll / 2)))
	cp := float32(math.Cos(float64(pitch / 2)))
	cy := float32(math.Cos(float64(yaw / 2)))
	sr := float32(math.Sin(float64(roll / 2)))
	sp := float32(math.Sin(float64(pitch / 2)))
	sy := float32(math.Sin(float64(yaw / 2)))

	return Quaternion{
		X: sr*cp*cy - cr*sp*sy,
		Y: cr*sp*cy + sr*cp*sy,
		Z: cr*cp*sy - sr*sp*cy,
		W: cr*cp*cy + sr*sp*sy,
	}
}

// QuaternionToEuler 四元数转欧拉角
func (q Quaternion) QuaternionToEuler() (roll, pitch, yaw float32) {
	// 计算pitch
	sinr_cosp := 2 * (q.W*q.X + q.Y*q.Z)
	cosr_cosp := 1 - 2*(q.X*q.X+q.Y*q.Y)
	roll = float32(math.Atan2(float64(sinr_cosp), float64(cosr_cosp)))

	// 计算pitch
	sinp := 2 * (q.W*q.Y - q.Z*q.X)
	if math.Abs(float64(sinp)) >= 1 {
		pitch = float32(math.Copysign(math.Pi/2, float64(sinp)))
	} else {
		pitch = float32(math.Asin(float64(sinp)))
	}

	// 计算yaw
	siny_cosp := 2 * (q.W*q.Z + q.X*q.Y)
	cosy_cosp := 1 - 2*(q.Y*q.Y+q.Z*q.Z)
	yaw = float32(math.Atan2(float64(siny_cosp), float64(cosy_cosp)))

	return
}

// Collider 碰撞盒
type Collider struct {
	Type    string  // 碰撞盒类型: sphere, box
	Radius  float32 // 球体半径
	Width   float32 // 盒子宽度
	Height  float32 // 盒子高度
	Depth   float32 // 盒子深度
	Offset  Vector3 // 碰撞盒偏移
	Trigger bool    // 是否为触发器
}

// NewSphereCollider 创建球体碰撞盒
func NewSphereCollider(radius float32, offset Vector3, trigger bool) Collider {
	return Collider{
		Type:    "sphere",
		Radius:  radius,
		Offset:  offset,
		Trigger: trigger,
	}
}

// NewBoxCollider 创建盒子碰撞盒
func NewBoxCollider(width, height, depth float32, offset Vector3, trigger bool) Collider {
	return Collider{
		Type:    "box",
		Width:   width,
		Height:  height,
		Depth:   depth,
		Offset:  offset,
		Trigger: trigger,
	}
}
