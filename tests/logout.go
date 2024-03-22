// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"

	"github.com/runfinch/common-tests/fnet"

	"github.com/runfinch/common-tests/command"
	"github.com/runfinch/common-tests/ffs"
	"github.com/runfinch/common-tests/option"
)

// Logout tests logging out a container registry.
func Logout(o *option.Option) {
	ginkgo.Describe("log out a container registry", func() {
		ginkgo.BeforeEach(func() {
			command.RemoveAll(o)
		})
		ginkgo.AfterEach(func() {
			command.RemoveAll(o)
		})
		ginkgo.When("the private registry is running and an image is built", func() {
			var registry string
			var tag string
			ginkgo.BeforeEach(func() {
				filename := "htpasswd"
				// The htpasswd is generated by
				// `<subject> run --entrypoint htpasswd public.ecr.aws/docker/library/httpd:2 -Bbn testUser testPassword`.
				// We don't want to generate it on the fly because:
				// 1. Pulling the httpd image can take a long time, sometimes even more 10 seconds.
				// 2. It's unlikely that we will have to update this in the future.
				// 3. It's not the thing we want to validate by the functional tests. We only want the output produced by it.
				//nolint:gosec // This password is only used for testing purpose.
				htpasswd := "testUser:$2y$05$wE0sj3r9O9K9q7R0MXcfPuIerl/06L1IsxXkCuUr3QZ8lHWwicIdS"
				htpasswdDir := filepath.Dir(ffs.CreateTempFile(filename, htpasswd))
				ginkgo.DeferCleanup(os.RemoveAll, htpasswdDir)
				port := fnet.GetFreePort()
				command.Run(o, "run",
					"-dp", fmt.Sprintf("%d:5000", port),
					"--name", "registry",
					"-v", fmt.Sprintf("%s:/auth", htpasswdDir),
					"-e", "REGISTRY_AUTH=htpasswd",
					"-e", "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm",
					"-e", fmt.Sprintf("REGISTRY_AUTH_HTPASSWD_PATH=/auth/%s", filename),
					registryImage)
				registry = fmt.Sprintf(`localhost:%d`, port)
				tag = fmt.Sprintf(`%s/test-login:tag`, registry)
				buildContext := ffs.CreateBuildContext(fmt.Sprintf(`FROM %s
		CMD ["echo", "bar"]
			`, localImages[defaultImage]))
				ginkgo.DeferCleanup(os.RemoveAll, buildContext)
				command.Run(o, "build", "-t", tag, buildContext)
			})
			ginkgo.It("should fail to push an image after logging out the registry", func() {
				command.Run(o, "login", registry, "-u", testUser, "-p", testPassword)
				ginkgo.DeferCleanup(func() {
					command.Run(o, "logout", registry)
				})
				command.Run(o, "push", tag)
				command.Run(o, "logout", registry)
				command.RunWithoutSuccessfulExit(o, "push", tag)
			})
		})
	})
}
