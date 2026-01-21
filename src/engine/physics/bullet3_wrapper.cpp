#include "bullet3_wrapper.h"
#include "btBulletDynamicsCommon.h"

extern "C" {

////////////////////////////////////////////////////////////////////////////////
// btBroadphaseInterface                                                      //
////////////////////////////////////////////////////////////////////////////////
btBroadphaseInterface* new_btBroadphaseInterface() {
	return new btDbvtBroadphase();
}

void destroy_btBroadphaseInterface(btBroadphaseInterface* broadphase) {
	delete broadphase;
}

////////////////////////////////////////////////////////////////////////////////
// btDefaultCollisionConfiguration                                            //
////////////////////////////////////////////////////////////////////////////////
btDefaultCollisionConfiguration* new_btDefaultCollisionConfiguration() {
	return new btDefaultCollisionConfiguration();
}

void destroy_btDefaultCollisionConfiguration(
	btDefaultCollisionConfiguration* config)
{
	delete config;
}

////////////////////////////////////////////////////////////////////////////////
// btCollisionDispatcher                                                      //
////////////////////////////////////////////////////////////////////////////////
btCollisionDispatcher* new_btCollisionDispatcher(
	btDefaultCollisionConfiguration* collisionConfig)
{
	return new btCollisionDispatcher(collisionConfig);
}

void destroy_btCollisionDispatcher(btCollisionDispatcher* dispatcher) {
	delete dispatcher;
}

////////////////////////////////////////////////////////////////////////////////
// btSequentialImpulseConstraintSolver                                        //
////////////////////////////////////////////////////////////////////////////////
btSequentialImpulseConstraintSolver* new_btSequentialImpulseConstraintSolver() {
	return new btSequentialImpulseConstraintSolver();
}

void destroy_btSequentialImpulseConstraintSolver(
	btSequentialImpulseConstraintSolver* solver)
{
	delete solver;
}

////////////////////////////////////////////////////////////////////////////////
// btDiscreteDynamicsWorld                                                    //
////////////////////////////////////////////////////////////////////////////////
btDiscreteDynamicsWorld* new_btDiscreteDynamicsWorld(
	btCollisionDispatcher* dispatcher,
	btBroadphaseInterface* broadphase,
	btSequentialImpulseConstraintSolver* solver,
	btDefaultCollisionConfiguration* collisionConfig)
{
	return new btDiscreteDynamicsWorld(dispatcher, broadphase, solver, collisionConfig);
}

void destroy_btDiscreteDynamicsWorld(btDiscreteDynamicsWorld* world) {
	delete world;
}

void btDiscreteDynamicsWorld_setGravity(
	btDiscreteDynamicsWorld* world,
	float x, float y, float z)
{
	world->setGravity(btVector3(x, y, z));
}

void btDiscreteDynamicsWorld_stepSimulation(
	btDiscreteDynamicsWorld* world,
	float timeStep)
{
	world->stepSimulation(timeStep, 10);
}

void btDiscreteDynamicsWorld_addRigidBody(
	btDiscreteDynamicsWorld* world, btRigidBody* body)
{
	world->addRigidBody(body);
}

void btDiscreteDynamicsWorld_removeRigidBody(
	btDiscreteDynamicsWorld* world, btRigidBody* body)
{
	world->removeRigidBody(body);
}

HitWrapper btDiscreteDynamicsWorld_rayTest(
	btDiscreteDynamicsWorld* world, float fx, float fy, float fz,
	float tx, float ty, float tz)
{
	btVector3 from(fx, fy, fz);
	btVector3 to(tx, ty, tz);
	btCollisionWorld::ClosestRayResultCallback rayCallback(from, to);
	rayCallback.m_collisionFilterGroup = btBroadphaseProxy::DefaultFilter;
	rayCallback.m_collisionFilterMask = btBroadphaseProxy::AllFilter;
	world->rayTest(from, to, rayCallback);
	HitWrapper hit = {};
	if (rayCallback.hasHit()) {
		memcpy(hit.point, rayCallback.m_hitPointWorld, sizeof(hit.point));
		memcpy(hit.normal, rayCallback.m_hitNormalWorld, sizeof(hit.normal));
		hit.object = rayCallback.m_collisionObject;
	}
	return hit;
}

HitWrapper btDiscreteDynamicsWorld_sphereSweep(
	btDiscreteDynamicsWorld* world, float fx, float fy, float fz,
	float tx, float ty, float tz, float radius)
{
	btVector3 from(fx, fy, fz);
	btVector3 to(tx, ty, tz);
    btSphereShape sphereShape(radius);
    btTransform fromTransform;
    fromTransform.setIdentity();
    fromTransform.setOrigin(from);
    btTransform toTransform;
    toTransform.setIdentity();
    toTransform.setOrigin(to);
    btCollisionWorld::ClosestConvexResultCallback sweepCallback(from, to);
    sweepCallback.m_collisionFilterGroup = btBroadphaseProxy::DefaultFilter;
    sweepCallback.m_collisionFilterMask = btBroadphaseProxy::AllFilter;
    world->convexSweepTest(&sphereShape, fromTransform, toTransform, sweepCallback);
	HitWrapper hit = {};
    if (sweepCallback.hasHit()) {
        memcpy(hit.point, sweepCallback.m_hitPointWorld, sizeof(hit.point));
		memcpy(hit.normal, sweepCallback.m_hitNormalWorld, sizeof(hit.normal));
		hit.object = sweepCallback.m_hitCollisionObject;
    }
	return hit;
}

////////////////////////////////////////////////////////////////////////////////
// btRigidBody                                                                //
////////////////////////////////////////////////////////////////////////////////
btRigidBody* new_btRigidBody(float mass,
	btMotionState* motion, btCollisionShape* shape,
	float localInertiaX, float localInertiaY, float localInertiaZ)
{
	btRigidBody::btRigidBodyConstructionInfo groundInfo(mass, motion, shape,
		btVector3(localInertiaX, localInertiaY, localInertiaZ));
	return new btRigidBody(groundInfo);
}

void destroy_btRigidBody(btRigidBody* body) { delete body; }

void btRigidBody_getPosition(btRigidBody* body, float* x, float* y, float* z) {
	btTransform trans;
	body->getMotionState()->getWorldTransform(trans);
	auto& o = trans.getOrigin();
	*x = o.getX();
	*y = o.getY();
	*z = o.getZ();
}

void btRigidBody_getRotation(btRigidBody* body, float* x, float* y, float* z, float* w) {
	btTransform trans;
	body->getMotionState()->getWorldTransform(trans);
	auto q = trans.getRotation();
	*x = q.getX();
	*y = q.getY();
	*z = q.getZ();
	*w = q.getW();
}

void btRigidBody_applyForceAtPoint(btRigidBody* body,
	float fx, float fy, float fz, float px, float py, float pz)
{
	btVector3 force(fx, fy, fz);
	btVector3 point(px, py, pz);
    body->activate(true);
    body->applyForce(force, point);
}

void btRigidBody_applyImpulseAtPoint(btRigidBody* body,
	float fx, float fy, float fz, float px, float py, float pz)
{
	btVector3 force(fx, fy, fz);
	btVector3 point(px, py, pz);
    body->activate(true);
    body->applyImpulse(force, point);
}
	
////////////////////////////////////////////////////////////////////////////////
// btCollisionShape                                                           //
////////////////////////////////////////////////////////////////////////////////
void btCollisionShape_calculateLocalInertia(btCollisionShape* shape,
	float mass, float* x, float* y, float* z)
{
	btVector3 inertia(0, 0, 0);
	shape->calculateLocalInertia(mass, inertia);
	*x = inertia.getX();
	*y = inertia.getY();
	*z = inertia.getZ();
}

void destroy_btCollisionShape(btCollisionShape* shape) { delete shape; }

btBoxShape* new_btBoxShape(float width, float height, float depth) {
	return new btBoxShape(btVector3(width, height, depth));
}

btSphereShape* new_btSphereShape(float radius) {
	return new btSphereShape(radius);
}

btCapsuleShape* new_btCapsuleShape(float radius, float height) {
	return new btCapsuleShape(radius, height);
}

btCylinderShape* new_btCylinderShape(float halfExtentsX, float halfExtentsY, float halfExtentsZ) {
	return new btCylinderShape(btVector3(halfExtentsX, halfExtentsY, halfExtentsZ));
}

btConeShape* new_btConeShape(float radius, float height) {
	return new btConeShape(radius, height);
}

btStaticPlaneShape* new_btStaticPlaneShape(float normalX, float normalY, float normalZ, float constant) {
	return new btStaticPlaneShape(btVector3(normalX, normalY, normalZ), constant);
}

btCompoundShape* new_btCompoundShape(int initialChildCapacity, bool enableDynamicAABBTree) {
	return new btCompoundShape(enableDynamicAABBTree, initialChildCapacity);
}

btConvexHullShape* new_btConvexHullShape(float* points, int numPoints, int stride) {
	return new btConvexHullShape(points, numPoints, stride);
}

// TODO:  Implement the following
//btBvhTriangleMeshShape* new_btBvhTriangleMeshShape(float* points, int numPoints, int stride) {
//	return new btBvhTriangleMeshShape();
//}

btEmptyShape* new_btEmptyShape() {
	return new btEmptyShape();
}

btMultiSphereShape* new_btMultiSphereShape(float* positions, float* radii, int numSpheres) {
	btVector3* vPositions = (btVector3*)malloc(numSpheres*sizeof(btVector3));
	for (int i = 0; i < numSpheres; i++) {
		vPositions->setX(positions[i*3]);
		vPositions->setY(positions[i*3+1]);
		vPositions->setZ(positions[i*3+2]);
	}
	return new btMultiSphereShape(vPositions, radii, numSpheres);
	free(vPositions);
}

btUniformScalingShape* new_btUniformScalingShape(btConvexShape* convexChildShape, float scaleFactor) {
	return new btUniformScalingShape(convexChildShape, scaleFactor);
}

////////////////////////////////////////////////////////////////////////////////
// btDefaultMotionState                                                       //
////////////////////////////////////////////////////////////////////////////////
btDefaultMotionState* new_btDefaultMotionState(
	float rotX, float rotY, float rotZ, float rotW,
	float centerMassX, float centerMassY, float centerMassZ)
{
	return new btDefaultMotionState(btTransform(
		btQuaternion(rotX, rotY, rotZ, rotW),
		btVector3(centerMassX, centerMassY, centerMassZ)));
}

void destroy_btMotionState(btMotionState* state) { delete state; }

} //extern "C" {
