package tracing

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// StartSpan 开始一个新的span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("jh_game_service")
	return tracer.Start(ctx, name, opts...)
}

// InitJaeger 初始化Jaeger追踪
func InitJaeger() (func(), error) {
	ctx := context.Background()

	// 从配置获取Jaeger设置
	jaegerConfig := g.Cfg().MustGet(ctx, "jaeger")
	if jaegerConfig.IsEmpty() {
		g.Log().Warning(ctx, "Jaeger配置为空，跳过初始化")
		return func() {}, nil
	}

	enabled := g.Cfg().MustGet(ctx, "jaeger.enabled", true).Bool()
	if !enabled {
		g.Log().Info(ctx, "Jaeger追踪已禁用")
		return func() {}, nil
	}

	endpoint := g.Cfg().MustGet(ctx, "jaeger.endpoint", "http://jaeger:14268/api/traces").String()
	serviceName := g.Cfg().MustGet(ctx, "jaeger.serviceName", "jh_game_service").String()
	serviceVersion := g.Cfg().MustGet(ctx, "jaeger.serviceVersion", "1.0.0").String()
	environment := g.Cfg().MustGet(ctx, "jaeger.environment", "development").String()

	g.Log().Infof(ctx, "初始化Jaeger追踪: endpoint=%s, service=%s", endpoint, serviceName)

	// 创建Jaeger导出器
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		return nil, fmt.Errorf("创建Jaeger导出器失败: %w", err)
	}

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// 创建追踪提供者
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()), // 开发环境采样所有请求
	)

	// 设置全局追踪提供者
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	g.Log().Info(ctx, "Jaeger追踪初始化成功")

	// 返回清理函数
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			g.Log().Errorf(ctx, "关闭追踪提供者失败: %v", err)
		} else {
			g.Log().Info(ctx, "Jaeger追踪已关闭")
		}
	}, nil
}
