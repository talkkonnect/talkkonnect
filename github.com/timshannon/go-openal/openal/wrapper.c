// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "local.h"
#include "wrapper.h"

const char *walGetString(ALenum param) {
	return alGetString(param);
}

void walGetBooleanv(ALenum param, void* data) {
	alGetBooleanv(param, data);
}

void walGetIntegerv(ALenum param, void* data) {
	alGetIntegerv(param, data);
}

void walGetFloatv(ALenum param, void* data) {
	alGetFloatv(param, data);
}

void walGetDoublev(ALenum param, void* data) {
	alGetDoublev(param, data);
}

// Listeners

void walListenerfv(ALenum param, const void* values) {
	alListenerfv(param, values);
}

void walListeneriv(ALenum param, const void* values) {
	alListeneriv(param, values);
}

ALfloat walGetListenerf(ALenum param) {
	ALfloat result;
	alGetListenerf(param, &result);
	return result;
}

void walGetListener3f(ALenum param, void *value1, void *value2, void *value3) {
	alGetListener3f(param, value1, value2, value3);
}

void walGetListenerfv(ALenum param, void* values) {
	alGetListenerfv(param, values);
}

ALint walGetListeneri(ALenum param) {
	ALint result;
	alGetListeneri(param, &result);
	return result;
}

void walGetListener3i(ALenum param, void *value1, void *value2, void *value3) {
	alGetListener3i(param, value1, value2, value3);
}

void walGetListeneriv(ALenum param, void* values) {
	alGetListeneriv(param, values);
}

// Sources

void walGenSources(ALsizei n, void *sources) {
	alGenSources(n, sources);
}

void walDeleteSources(ALsizei n, const void *sources) {
	alDeleteSources(n, sources);
}

void walSourcefv(ALuint sid, ALenum param, const void* values) {
	alSourcefv(sid, param, values);
}

void walSourceiv(ALuint sid, ALenum param, const void* values) {
	alSourceiv(sid, param, values);
}

ALfloat walGetSourcef(ALuint sid, ALenum param) {
	ALfloat result;
	alGetSourcef(sid, param, &result);
	return result;
}

void walGetSource3f(ALuint sid, ALenum param, void *value1, void *value2, void *value3) {
	alGetSource3f(sid, param, value1, value2, value3);
}

void walGetSourcefv(ALuint sid, ALenum param, void* values) {
	alGetSourcefv(sid, param, values);
}

ALint walGetSourcei(ALuint sid, ALenum param) {
	ALint result;
	alGetSourcei(sid, param, &result);
	return result;
}

void walGetSource3i(ALuint sid, ALenum param, void *value1, void *value2, void *value3) {
	alGetSource3i(sid, param, value1, value2, value3);
}

void walGetSourceiv(ALuint sid, ALenum param, void* values) {
	alGetSourceiv(sid, param, values);
}

void walSourcePlayv(ALsizei ns, const void *sids) {
	alSourcePlayv(ns, sids);
}

void walSourceStopv(ALsizei ns, const void *sids) {
	alSourceStopv(ns, sids);
}

void walSourceRewindv(ALsizei ns, const void *sids) {
	alSourceRewindv(ns, sids);
}

void walSourcePausev(ALsizei ns, const void *sids) {
	alSourcePausev(ns, sids);
}

void walSourceQueueBuffers(ALuint sid, ALsizei numEntries, const void *bids) {
	alSourceQueueBuffers(sid, numEntries, bids);
}

void walSourceUnqueueBuffers(ALuint sid, ALsizei numEntries, void *bids) {
	alSourceUnqueueBuffers(sid, numEntries, bids);
}

// Buffers

void walGenBuffers(ALsizei n, void *buffers) {
	alGenBuffers(n, buffers);
}

void walDeleteBuffers(ALsizei n, const void *buffers) {
	alDeleteBuffers(n, buffers);
}

void walBufferfv(ALuint bid, ALenum param, const void* values) {
	alBufferfv(bid, param, values);
}

void walBufferiv(ALuint bid, ALenum param, const void* values) {
	alBufferiv(bid, param, values);
}

ALfloat walGetBufferf(ALuint bid, ALenum param) {
	ALfloat result;
	alGetBufferf(bid, param, &result);
	return result;
}

void walGetBuffer3f(ALuint bid, ALenum param, void *value1, void *value2, void *value3) {
	alGetBuffer3f(bid, param, value1, value2, value3);
}

void walGetBufferfv(ALuint bid, ALenum param, void* values) {
	alGetBufferfv(bid, param, values);
}

ALint walGetBufferi(ALuint bid, ALenum param) {
	ALint result;
	alGetBufferi(bid, param, &result);
	return result;
}

void walGetBuffer3i(ALuint bid, ALenum param, void *value1, void *value2, void *value3) {
	alGetBuffer3i(bid, param, value1, value2, value3);
}

void walGetBufferiv(ALuint bid, ALenum param, void* values) {
	alGetBufferiv(bid, param, values);
}

// Singulars

ALuint walGenSource(void) {
	ALuint source;
	alGenSources(1, &source);
	return source;
}

void walDeleteSource(ALuint source) {
	alDeleteSources(1, &source);
}

ALuint walGenBuffer(void) {
	ALuint buffer;
	alGenBuffers(1, &buffer);
	return buffer;
}

void walDeleteBuffer(ALuint buffer) {
	alDeleteBuffers(1, &buffer);
}

void walSourceQueueBuffer(ALuint sid, ALuint bid) {
	alSourceQueueBuffers(sid, 1, &bid);
}

ALuint walSourceUnqueueBuffer(ALuint sid) {
	ALuint result;
	alSourceUnqueueBuffers(sid, 1, &result);
	return result;
}
