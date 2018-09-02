#ifndef _GO_WRAPPER_AL_
#define _GO_WRAPPER_AL_

// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// It's sad but the OpenAL C API uses lots and lots of typedefs
// that require wrapper functions (using basic C types) for cgo
// to grok them. So there's a lot more C code here than I would
// like...

const char *walGetString(ALenum param);
void walGetBooleanv(ALenum param, void* data);
void walGetIntegerv(ALenum param, void* data);
void walGetFloatv(ALenum param, void* data);
void walGetDoublev(ALenum param, void* data);

// We don't define wrappers for these because we have
// no clue how to make Go grok C function pointers at
// runtime. So for now, OpenAL extensions can not be
// used from Go. If you have an idea for how to make
// it work, be sure to email! I suspect we'd need a
// mechanism for generating cgo-style stubs at runtime,
// sounds like work.
//
// ALboolean alIsExtensionPresent( const ALchar* extname );
// void* alGetProcAddress( const ALchar* fname );
// ALenum alGetEnumValue( const ALchar* ename );

// Listeners

void walListenerfv(ALenum param, const void* values);
void walListeneriv(ALenum param, const void* values);
ALfloat walGetListenerf(ALenum param);
void walGetListener3f(ALenum param, void *value1, void *value2, void *value3);
void walGetListenerfv(ALenum param, void* values);
ALint walGetListeneri(ALenum param);
void walGetListener3i(ALenum param, void *value1, void *value2, void *value3);
void walGetListeneriv(ALenum param, void* values);

// Sources

void walGenSources(ALsizei n, void *sources);
void walDeleteSources(ALsizei n, const void *sources);
void walSourcefv(ALuint sid, ALenum param, const void* values);
void walSourceiv(ALuint sid, ALenum param, const void* values);
ALfloat walGetSourcef(ALuint sid, ALenum param);
void walGetSource3f(ALuint sid, ALenum param, void *value1, void *value2, void *value3);
void walGetSourcefv(ALuint sid, ALenum param, void* values);
ALint walGetSourcei(ALuint sid, ALenum param);
void walGetSource3i(ALuint sid, ALenum param, void *value1, void *value2, void *value3);
void walGetSourceiv(ALuint sid, ALenum param, void* values);
void walSourcePlayv(ALsizei ns, const void *sids);
void walSourceStopv(ALsizei ns, const void *sids);
void walSourceRewindv(ALsizei ns, const void *sids);
void walSourcePausev(ALsizei ns, const void *sids);
void walSourceQueueBuffers(ALuint sid, ALsizei numEntries, const void *bids);
void walSourceUnqueueBuffers(ALuint sid, ALsizei numEntries, void *bids);

// Buffers

void walGenBuffers(ALsizei n, void *buffers);
void walDeleteBuffers(ALsizei n, const void *buffers);
void walBufferfv(ALuint bid, ALenum param, const void* values);
void walBufferiv(ALuint bid, ALenum param, const void* values);
ALfloat walGetBufferf(ALuint bid, ALenum param);
void walGetBuffer3f(ALuint bid, ALenum param, void *value1, void *value2, void *value3);
void walGetBufferfv(ALuint bid, ALenum param, void* values);
ALint walGetBufferi(ALuint bid, ALenum param);
void walGetBuffer3i(ALuint bid, ALenum param, void *value1, void *value2, void *value3);
void walGetBufferiv(ALuint bid, ALenum param, void* values);

// For convenience we offer "singular" versions of the following
// calls as well, which require different wrappers if we want to
// be efficient. The main reason for "singular" versions is that
// Go doesn't allow us to treat a variable as an array of size 1.

ALuint walGenSource(void);
void walDeleteSource(ALuint source);
ALuint walGenBuffer(void);
void walDeleteBuffer(ALuint buffer);
void walSourceQueueBuffer(ALuint sid, ALuint bid);
ALuint walSourceUnqueueBuffer(ALuint sid);

#endif
