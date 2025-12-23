#ifndef BULLET3_WRAPPER_H
#define BULLET3_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

typedef struct btBroadphaseInterface btBroadphaseInterface;
typedef struct btDefaultCollisionConfiguration btDefaultCollisionConfiguration;
typedef struct btCollisionDispatcher btCollisionDispatcher;
typedef struct btSequentialImpulseConstraintSolver btSequentialImpulseConstraintSolver;
typedef struct btDiscreteDynamicsWorld btDiscreteDynamicsWorld;
typedef struct btRigidBody btRigidBody;
typedef struct btCollisionShape btCollisionShape;
typedef struct btBoxShape btBoxShape;
typedef struct btSphereShape btSphereShape;
typedef struct btCapsuleShape btCapsuleShape;
typedef struct btCylinderShape btCylinderShape;
typedef struct btConeShape btConeShape;
typedef struct btStaticPlaneShape btStaticPlaneShape;
typedef struct btCompoundShape btCompoundShape;
typedef struct btConvexShape btConvexShape;
typedef struct btConvexHullShape btConvexHullShape;
typedef struct btBvhTriangleMeshShape btBvhTriangleMeshShape;
typedef struct btHeightfieldTerrainShape btHeightfieldTerrainShape;
typedef struct btTetrahedralShape btTetrahedralShape;
typedef struct btEmptyShape btEmptyShape;
typedef struct btMultiSphereShape btMultiSphereShape;
typedef struct btUniformScalingShape btUniformScalingShape;
typedef struct btMotionState btMotionState;
typedef struct btDefaultMotionState btDefaultMotionState;
typedef struct btCollisionObject btCollisionObject;

typedef struct {
	float point[4];
	float normal[4];
	const btCollisionObject* object;
} HitWrapper;

////////////////////////////////////////////////////////////////////////////////
// btBroadphaseInterface                                                      //
////////////////////////////////////////////////////////////////////////////////
btBroadphaseInterface* new_btBroadphaseInterface();
void destroy_btBroadphaseInterface(btBroadphaseInterface* broadphase);

////////////////////////////////////////////////////////////////////////////////
// btDefaultCollisionConfiguration                                            //
////////////////////////////////////////////////////////////////////////////////
btDefaultCollisionConfiguration* new_btDefaultCollisionConfiguration();
void destroy_btDefaultCollisionConfiguration(
	btDefaultCollisionConfiguration* config);

////////////////////////////////////////////////////////////////////////////////
// btCollisionDispatcher                                                      //
////////////////////////////////////////////////////////////////////////////////
btCollisionDispatcher* new_btCollisionDispatcher(
    btDefaultCollisionConfiguration* collisionConfig);
void destroy_btCollisionDispatcher(btCollisionDispatcher* dispatcher);

////////////////////////////////////////////////////////////////////////////////
// btSequentialImpulseConstraintSolver                                        //
////////////////////////////////////////////////////////////////////////////////
btSequentialImpulseConstraintSolver* new_btSequentialImpulseConstraintSolver();
void destroy_btSequentialImpulseConstraintSolver(
	btSequentialImpulseConstraintSolver* solver);

////////////////////////////////////////////////////////////////////////////////
// btDiscreteDynamicsWorld                                                    //
////////////////////////////////////////////////////////////////////////////////
btDiscreteDynamicsWorld* new_btDiscreteDynamicsWorld(
	btCollisionDispatcher* dispatcher,
	btBroadphaseInterface* broadphase,
	btSequentialImpulseConstraintSolver* solver,
	btDefaultCollisionConfiguration* collisionConfig);
void destroy_btDiscreteDynamicsWorld(btDiscreteDynamicsWorld* world);
void btDiscreteDynamicsWorld_setGravity(
	btDiscreteDynamicsWorld* world,
	float x, float y, float z);
void btDiscreteDynamicsWorld_stepSimulation(
	btDiscreteDynamicsWorld* world,
	float timeStep);
void btDiscreteDynamicsWorld_addRigidBody(
	btDiscreteDynamicsWorld* world, btRigidBody* body);
void btDiscreteDynamicsWorld_removeRigidBody(
	btDiscreteDynamicsWorld* world, btRigidBody* body);
HitWrapper btDiscreteDynamicsWorld_rayTest(
	btDiscreteDynamicsWorld* world, float fx, float fy, float fz,
	float tx, float ty, float tz);
HitWrapper btDiscreteDynamicsWorld_sphereSweep(
	btDiscreteDynamicsWorld* world, float fx, float fy, float fz,
	float tx, float ty, float tz, float radius);

////////////////////////////////////////////////////////////////////////////////
// btRigidBody                                                                //
////////////////////////////////////////////////////////////////////////////////
btRigidBody* new_btRigidBody(float mass,
	btMotionState* motion, btCollisionShape* shape,
	float localInertiaX, float localInertiaY, float localInertiaZ);
void destroy_btRigidBody(btRigidBody* body);
void btRigidBody_getPosition(btRigidBody* body, float* x, float* y, float* z);
void btRigidBody_getRotation(btRigidBody* body, float* x, float* y, float* z, float* w);
void btRigidBody_applyForceAtPoint(btRigidBody* body,
	float fx, float fy, float fz, float px, float py, float pz);
void btRigidBody_applyImpulseAtPoint(btRigidBody* body,
	float fx, float fy, float fz, float px, float py, float pz);

////////////////////////////////////////////////////////////////////////////////
// btCollisionShape                                                           //
////////////////////////////////////////////////////////////////////////////////
void btCollisionShape_calculateLocalInertia(btCollisionShape* shape,
	float mass, float* x, float* y, float* z);
void destroy_btCollisionShape(btCollisionShape* shape);
btBoxShape* new_btBoxShape(float width, float height, float depth);
btSphereShape* new_btSphereShape(float radius);
btCapsuleShape* new_btCapsuleShape(float radius, float height);
btCylinderShape* new_btCylinderShape(float halfExtentsX, float halfExtentsY, float halfExtentsZ);
btConeShape* new_btConeShape(float radius, float height);
btStaticPlaneShape* new_btStaticPlaneShape(float normalX, float normalY, float normalZ, float constant);
btCompoundShape* new_btCompoundShape(int initialChildCapacity, bool enableDynamicAABBTree);
btConvexHullShape* new_btConvexHullShape(float* points, int numPoints, int stride);
btEmptyShape* new_btEmptyShape();
btMultiSphereShape* new_btMultiSphereShape(float* positions, float* radii, int numSpheres);
btUniformScalingShape* new_btUniformScalingShape(btConvexShape* convexChildShape, float scaleFactor);

////////////////////////////////////////////////////////////////////////////////
// btDefaultMotionState                                                       //
////////////////////////////////////////////////////////////////////////////////
btDefaultMotionState* new_btDefaultMotionState(
	float rotX, float rotY, float rotZ, float rotW,
	float centerMassX, float centerMassY, float centerMassZ);
void destroy_btMotionState(btMotionState* state);

#ifdef __cplusplus
}
#endif

#endif
