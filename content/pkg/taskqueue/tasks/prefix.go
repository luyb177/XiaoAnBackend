package tasks

type TaskPrefix string

const (
	ArticleRelationTaskPrefix TaskPrefix = "article_relation_task"
	VideoRelationTaskPrefix   TaskPrefix = "video_relation_task"
	PodcastRelationTaskPrefix TaskPrefix = "podcast_relation_task"
)
