package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	proj := flag.String("project", "", "GCP project ID")
	topicID := flag.String("topic", "", "Cloud Pub/Sub topic")
	flag.Parse()

	if err := requireString("project", proj); err != nil {
		return err
	}

	if flag.NArg() == 0 {
		return errors.New("must specify resouce: [topic, subscription]")
	}

	client, err := pubsub.NewClient(ctx, *proj)
	if err != nil {
		return err
	}
	defer client.Close()

	switch res := flag.Arg(0); res {
	case "topic":
		if flag.NArg() < 2 {
			return errors.New("must specify subcommand: [create]")
		}
		switch cmd := flag.Arg(1); cmd {
		case "create":
			name := flag.Arg(2)
			if name == "" {
				return errors.New("must specify a topic name")
			}
			_, err := client.CreateTopic(ctx, name)
			if err != nil {
				return err
			}
		case "publish":
			body := flag.Arg(2)
			if body == "" {
				return errors.New("must specify a message body")
			}
			if err := requireString("topic", topicID); err != nil {
				return err
			}
			topic, err := getTopic(ctx, *topicID, client)
			if err != nil {
				return err
			}
			_ = topic.Publish(ctx, &pubsub.Message{Data: []byte(body)})
			topic.Stop()
		default:
			return fmt.Errorf("unknown subcommand: %q for %s", cmd, res)
		}
	case "subscription":
		if flag.NArg() < 2 {
			return errors.New("must specify subcommand: [create]")
		}
		switch cmd := flag.Arg(1); cmd {
		case "create":
			name := flag.Arg(2)
			if name == "" {
				return errors.New("must specify a subscription name")
			}
			if err := requireString("topic", topicID); err != nil {
				return err
			}
			topic, err := getTopic(ctx, *topicID, client)
			if err != nil {
				return err
			}
			_, err = client.CreateSubscription(ctx, name, pubsub.SubscriptionConfig{Topic: topic})
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown subcommand: %q for %s", cmd, res)
		}
	default:
		return fmt.Errorf("unknown resource: %s", res)
	}

	return nil
}

func requireString(flag string, v *string) error {
	if v == nil || *v == "" {
		return fmt.Errorf("must specify --%s", flag)
	}
	return nil
}

func getTopic(ctx context.Context, id string, client *pubsub.Client) (*pubsub.Topic, error) {
	topic := client.Topic(id)
	if ok, err := topic.Exists(ctx); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("Topic(%q) does not exist", id)
	}
	return topic, nil
}
