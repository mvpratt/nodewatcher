package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.24

import (
	"context"
	"log"
	"time"

	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/graph/model"
)

// CreatedAt is the resolver for the created_at field.
func (r *multiChannelBackupResolver) CreatedAt(ctx context.Context, obj *model.MultiChannelBackup) (string, error) {
	return obj.CreatedAt.Format(time.RFC850), nil
}

// CreateNode is the resolver for the createNode field.
func (r *mutationResolver) CreateNode(ctx context.Context, input model.NewNode) (*model.Node, error) {
	node := &model.Node{
		ID:       int64(input.ID),
		URL:      input.URL,
		Alias:    input.Alias,
		Pubkey:   input.Pubkey,
		Macaroon: input.Macaroon,
	}

	dbNode := &db.Node{
		ID:       0,
		URL:      input.URL,
		Alias:    input.Alias,
		Pubkey:   input.Pubkey,
		Macaroon: input.Macaroon,
	}
	err := r.DB.InsertNode(dbNode)
	if err != nil {
		log.Print(err.Error())
	}
	return node, nil
}

// Nodes is the resolver for the nodes field.
func (r *queryResolver) Nodes(ctx context.Context) ([]*model.Node, error) {
	nodes, err := r.DB.FindAllNodes()
	if err != nil {
		return nil, err
	}

	var graphNodes []*model.Node

	var g *model.Node
	for _, node := range nodes {
		g = &model.Node{
			ID:       int64(node.ID),
			URL:      node.URL,
			Alias:    node.Alias,
			Pubkey:   node.Pubkey,
			Macaroon: node.Macaroon,
		}
		graphNodes = append(graphNodes, g)
	}
	return graphNodes, nil
}

// Channels is the resolver for the channels field.
func (r *queryResolver) Channels(ctx context.Context) ([]*model.Channel, error) {
	channels, err := r.DB.FindAllChannels()
	if err != nil {
		return nil, err
	}

	var graphChannels []*model.Channel

	var g *model.Channel
	for _, channel := range channels {
		g = &model.Channel{
			ID:          channel.ID,
			FundingTxid: channel.FundingTxid,
			OutputIndex: channel.OutputIndex,
			NodeID:      channel.NodeID,
		}
		graphChannels = append(graphChannels, g)
	}
	return graphChannels, nil
}

// MultiChannelBackups is the resolver for the multi_channel_backups field.
func (r *queryResolver) MultiChannelBackups(ctx context.Context) ([]*model.MultiChannelBackup, error) {
	channels, err := r.DB.FindAllMultiChannelBackups()
	if err != nil {
		return nil, err
	}

	var graphChannels []*model.MultiChannelBackup

	var g *model.MultiChannelBackup
	for _, channel := range channels {
		g = &model.MultiChannelBackup{
			ID:        channel.ID,
			CreatedAt: channel.CreatedAt,
			Backup:    channel.Backup,
			NodeID:    channel.NodeID,
		}
		graphChannels = append(graphChannels, g)
	}
	return graphChannels, nil
}

// MultiChannelBackup returns MultiChannelBackupResolver implementation.
func (r *Resolver) MultiChannelBackup() MultiChannelBackupResolver {
	return &multiChannelBackupResolver{r}
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type multiChannelBackupResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
