package decorator

import (
	"fmt"

	sparta "github.com/mweagle/Sparta"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

// CloudWatchErrorAlarmDecorator returns a TemplateDecoratorHookFunc
// that associates a CloudWatch Lambda Error count alarm with the given
// lambda function. The four parameters are periodWindow
func CloudWatchErrorAlarmDecorator(periodWindow int,
	minutesPerPeriod int,
	thresholdGreaterThanOrEqualToValue int,
	snsTopic gocf.Stringable) sparta.TemplateDecoratorHookFunc {
	alarmDecorator := func(serviceName string,
		lambdaResourceName string,
		lambdaResource gocf.LambdaFunction,
		resourceMetadata map[string]interface{},
		S3Bucket string,
		S3Key string,
		buildID string,
		template *gocf.Template,
		context map[string]interface{},
		logger *logrus.Logger) error {

		periodInSeconds := minutesPerPeriod * 60

		alarm := &gocf.CloudWatchAlarm{
			AlarmName: gocf.Join("",
				gocf.String("ERROR Alarm for "),
				gocf.Ref(lambdaResourceName)),
			AlarmDescription: gocf.Join(" ",
				gocf.String("ERROR count for AWS Lambda function"),
				gocf.Ref(lambdaResourceName),
				gocf.String("( Stack:"),
				gocf.Ref("AWS::StackName"),
				gocf.String(") is greater than"),
				gocf.String(fmt.Sprintf("%d", thresholdGreaterThanOrEqualToValue)),
				gocf.String("over the last"),
				gocf.String(fmt.Sprintf("%d", periodInSeconds)),
				gocf.String("seconds"),
			),
			MetricName:         gocf.String("Errors"),
			Namespace:          gocf.String("AWS/Lambda"),
			Statistic:          gocf.String("Sum"),
			Period:             gocf.Integer(int64(periodInSeconds)),
			EvaluationPeriods:  gocf.Integer(int64(periodWindow)),
			Threshold:          gocf.Integer(int64(thresholdGreaterThanOrEqualToValue)),
			ComparisonOperator: gocf.String("GreaterThanOrEqualToThreshold"),
			Dimensions: &gocf.CloudWatchAlarmDimensionList{
				gocf.CloudWatchAlarmDimension{
					Name:  gocf.String("FunctionName"),
					Value: gocf.Ref(lambdaResourceName).String(),
				},
			},
			TreatMissingData: gocf.String("notBreaching"),
			AlarmActions: gocf.StringList(
				snsTopic,
			),
		}
		// Create the resource, add it...
		alarmResourceName := sparta.CloudFormationResourceName("Alarm",
			lambdaResourceName)
		template.AddResource(alarmResourceName, alarm)
		return nil
	}
	return sparta.TemplateDecoratorHookFunc(alarmDecorator)
}
