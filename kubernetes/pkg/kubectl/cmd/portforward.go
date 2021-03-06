/*
Copyright 2014 The Kubernetes Authors.

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

package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"

	"github.com/openshift/github.com/spf13/cobra"

	"github.com/openshift/kubernetes/pkg/api"
	coreclient "github.com/openshift/kubernetes/pkg/client/clientset_generated/internalclientset/typed/core/internalversion"
	"github.com/openshift/kubernetes/pkg/client/restclient"
	"github.com/openshift/kubernetes/pkg/client/unversioned/portforward"
	"github.com/openshift/kubernetes/pkg/client/unversioned/remotecommand"
	"github.com/openshift/kubernetes/pkg/kubectl/cmd/templates"
	cmdutil "github.com/openshift/kubernetes/pkg/kubectl/cmd/util"
)

// PortForwardOptions contains all the options for running the port-forward cli command.
type PortForwardOptions struct {
	Namespace     string
	PodName       string
	RESTClient    *restclient.RESTClient
	Config        *restclient.Config
	PodClient     coreclient.PodsGetter
	Ports         []string
	PortForwarder portForwarder
	StopChannel   chan struct{}
	ReadyChannel  chan struct{}
}

var (
	portforward_example = templates.Examples(`
		# Listen on ports 5000 and 6000 locally, forwarding data to/from ports 5000 and 6000 in the pod
		kubectl port-forward mypod 5000 6000

		# Listen on port 8888 locally, forwarding to 5000 in the pod
		kubectl port-forward mypod 8888:5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		kubectl port-forward mypod :5000

		# Listen on a random port locally, forwarding to 5000 in the pod
		kubectl port-forward  mypod 0:5000`)
)

func NewCmdPortForward(f cmdutil.Factory, cmdOut, cmdErr io.Writer) *cobra.Command {
	opts := &PortForwardOptions{
		PortForwarder: &defaultPortForwarder{
			cmdOut: cmdOut,
			cmdErr: cmdErr,
		},
	}
	cmd := &cobra.Command{
		Use:     "port-forward POD [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]",
		Short:   "Forward one or more local ports to a pod",
		Long:    "Forward one or more local ports to a pod.",
		Example: portforward_example,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(f, cmd, args, cmdOut, cmdErr); err != nil {
				cmdutil.CheckErr(err)
			}
			if err := opts.Validate(); err != nil {
				cmdutil.CheckErr(cmdutil.UsageError(cmd, err.Error()))
			}
			if err := opts.RunPortForward(); err != nil {
				cmdutil.CheckErr(err)
			}
		},
	}
	cmd.Flags().StringP("pod", "p", "", "Pod name")
	// TODO support UID
	return cmd
}

type portForwarder interface {
	ForwardPorts(method string, url *url.URL, opts PortForwardOptions) error
}

type defaultPortForwarder struct {
	cmdOut, cmdErr io.Writer
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts PortForwardOptions) error {
	dialer, err := remotecommand.NewExecutor(opts.Config, method, url)
	if err != nil {
		return err
	}
	fw, err := portforward.New(dialer, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.cmdOut, f.cmdErr)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// Complete completes all the required options for port-forward cmd.
func (o *PortForwardOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string, cmdOut io.Writer, cmdErr io.Writer) error {
	var err error
	o.PodName = cmdutil.GetFlagString(cmd, "pod")
	if len(o.PodName) == 0 && len(args) == 0 {
		return cmdutil.UsageError(cmd, "POD is required for port-forward")
	}

	if len(o.PodName) != 0 {
		printDeprecationWarning("port-forward POD", "-p POD")
		o.Ports = args
	} else {
		o.PodName = args[0]
		o.Ports = args[1:]
	}

	o.Namespace, _, err = f.DefaultNamespace()
	if err != nil {
		return err
	}

	clientset, err := f.ClientSet()
	if err != nil {
		return err
	}
	o.PodClient = clientset.Core()

	o.Config, err = f.ClientConfig()
	if err != nil {
		return err
	}
	o.RESTClient, err = f.RESTClient()
	if err != nil {
		return err
	}

	o.StopChannel = make(chan struct{}, 1)
	o.ReadyChannel = make(chan struct{})
	return nil
}

// Validate validates all the required options for port-forward cmd.
func (o PortForwardOptions) Validate() error {
	if len(o.PodName) == 0 {
		return fmt.Errorf("pod name must be specified")
	}

	if len(o.Ports) < 1 {
		return fmt.Errorf("at least 1 PORT is required for port-forward")
	}

	if o.PortForwarder == nil || o.PodClient == nil || o.RESTClient == nil || o.Config == nil {
		return fmt.Errorf("client, client config, restClient, and portforwarder must be provided")
	}
	return nil
}

// RunPortForward implements all the necessary functionality for port-forward cmd.
func (o PortForwardOptions) RunPortForward() error {
	pod, err := o.PodClient.Pods(o.Namespace).Get(o.PodName)
	if err != nil {
		return err
	}

	if pod.Status.Phase != api.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		if o.StopChannel != nil {
			close(o.StopChannel)
		}
	}()

	req := o.RESTClient.Post().
		Resource("pods").
		Namespace(o.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return o.PortForwarder.ForwardPorts("POST", req.URL(), o)
}
