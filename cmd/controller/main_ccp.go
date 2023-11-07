//go:build ccp

/*
Portions Copyright (c) Microsoft Corporation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/samber/lo"

	"github.com/Azure/karpenter/pkg/cloudprovider"
	"github.com/Azure/karpenter/pkg/operator"

	altOperator "github.com/Azure/karpenter/pkg/alt/karpenter-core/pkg/operator"
	controllers "github.com/Azure/karpenter/pkg/controllers"
	"github.com/aws/karpenter-core/pkg/cloudprovider/metrics"
	corecontrollers "github.com/aws/karpenter-core/pkg/controllers"

	// Note the absence of corewebhooks: these pull in knative webhook-related packages and informers in init()
	// We don't give cluster-level roles when running in AKS managed mode, so their informers will produce errors and halt all other operations
	// corewebhooks "github.com/aws/karpenter-core/pkg/webhooks"

	"github.com/aws/karpenter-core/pkg/controllers/state"
)

func main() {
	//ctx, op := operator.NewOperator(coreoperator.NewOperator())
	ctx, op := operator.NewOperator(altOperator.NewOperator())
	aksCloudProvider := cloudprovider.New(
		op.InstanceTypesProvider,
		op.InstanceProvider,
		op.EventRecorder,
		op.GetClient(),
		op.ImageProvider,
	)

	lo.Must0(op.AddHealthzCheck("cloud-provider", aksCloudProvider.LivenessProbe))
	cloudProvider := metrics.Decorate(aksCloudProvider)

	op.
		WithControllers(ctx, corecontrollers.NewControllers(
			op.Clock,
			op.GetClient(),
			op.KubernetesInterface,
			state.NewCluster(op.Clock, op.GetClient(), cloudProvider),
			op.EventRecorder,
			cloudProvider,
		)...).
		// WithWebhooks(ctx, corewebhooks.NewWebhooks()...).
		WithControllers(ctx, controllers.NewControllers(
			ctx,
			op.GetClient(),
			aksCloudProvider,
			op.InstanceProvider,
		)...).
		// WithWebhooks(ctx, corewebhooks.NewWebhooks()...).
		Start(ctx)
}
