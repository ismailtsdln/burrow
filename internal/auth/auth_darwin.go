//go:build darwin

package auth

/*
#cgo LDFLAGS: -framework LocalAuthentication -framework Foundation
#include <LocalAuthentication/LocalAuthentication.h>
#include <Foundation/Foundation.h>
#include <stdlib.h>

extern void authCallback(int success);

static void authenticate(const char *reason) {
    LAContext *context = [[LAContext alloc] init];
    NSError *error = nil;
    NSString *nsReason = [NSString stringWithUTF8String:reason];

    LAPolicy policy = LAPolicyDeviceOwnerAuthentication;

    // Check if biometric authentication is available
    if ([context canEvaluatePolicy:LAPolicyDeviceOwnerAuthenticationWithBiometrics error:&error]) {
        policy = LAPolicyDeviceOwnerAuthenticationWithBiometrics;
    }

    [context evaluatePolicy:policy
            localizedReason:nsReason
                      reply:^(BOOL success, NSError * _Nullable error) {
        authCallback(success ? 1 : 0);
    }];
}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

var authChan = make(chan bool)

//export authCallback
func authCallback(success C.int) {
	authChan <- (success != 0)
}

type darwinAuthenticator struct{}

func (d *darwinAuthenticator) Authenticate(reason string) (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cReason := C.CString(reason)
	defer C.free(unsafe.Pointer(cReason))

	C.authenticate(cReason)

	success := <-authChan
	return success, nil
}

func getPlatformAuthenticator() Authenticator {
	return &darwinAuthenticator{}
}
