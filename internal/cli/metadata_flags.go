package cli

import (
	"github.com/spf13/cobra"

	"pdfmeta/internal/model"
)

type metadataFlagUsage int

const (
	metadataFlagUsageDefault metadataFlagUsage = iota
	metadataFlagUsageOverride
)

type metadataStringFlags struct {
	title      string
	author     string
	subject    string
	keywords   string
	creator    string
	producer   string
	createdAt  string
	modifiedAt string
}

func addMetadataPatchFlags(cmd *cobra.Command, f *metadataStringFlags, usage metadataFlagUsage) {
	flags := cmd.Flags()
	titleHelp := "Title"
	authorHelp := "Author"
	subjectHelp := "Subject"
	keywordsHelp := "Keywords"
	creatorHelp := "Creator"
	producerHelp := "Producer"
	creationDateHelp := "Creation date"
	modDateHelp := "Modification date"
	if usage == metadataFlagUsageOverride {
		titleHelp = "Override Title"
		authorHelp = "Override Author"
		subjectHelp = "Override Subject"
		keywordsHelp = "Override Keywords"
		creatorHelp = "Override Creator"
		producerHelp = "Override Producer"
		creationDateHelp = "Override creation date"
		modDateHelp = "Override modification date"
	}

	flags.StringVar(&f.title, "title", "", titleHelp)
	flags.StringVar(&f.author, "author", "", authorHelp)
	flags.StringVar(&f.subject, "subject", "", subjectHelp)
	flags.StringVar(&f.keywords, "keywords", "", keywordsHelp)
	flags.StringVar(&f.creator, "creator", "", creatorHelp)
	flags.StringVar(&f.producer, "producer", "", producerHelp)
	flags.StringVar(&f.createdAt, "creation-date", "", creationDateHelp)
	flags.StringVar(&f.modifiedAt, "mod-date", "", modDateHelp)
}

func patchFromMetadataFlags(cmd *cobra.Command, f *metadataStringFlags) model.MetadataPatch {
	var patch model.MetadataPatch
	if cmd.Flags().Changed("title") {
		patch.Title = &f.title
	}
	if cmd.Flags().Changed("author") {
		patch.Author = &f.author
	}
	if cmd.Flags().Changed("subject") {
		patch.Subject = &f.subject
	}
	if cmd.Flags().Changed("keywords") {
		patch.Keywords = &f.keywords
	}
	if cmd.Flags().Changed("creator") {
		patch.Creator = &f.creator
	}
	if cmd.Flags().Changed("producer") {
		patch.Producer = &f.producer
	}
	if cmd.Flags().Changed("creation-date") {
		patch.CreationDate = &f.createdAt
	}
	if cmd.Flags().Changed("mod-date") {
		patch.ModDate = &f.modifiedAt
	}
	return patch
}
