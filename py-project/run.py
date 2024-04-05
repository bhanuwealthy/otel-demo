import sys
import os

from fastapi import FastAPI
import requests


from opentelemetry import metrics, trace
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader  # , ConsoleMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor  # , ConsoleSpanExporter
from opentelemetry.semconv.resource import ResourceAttributes as ResAttr, TelemetrySdkLanguageValues as Lang
from opentelemetry.sdk.environment_variables import OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT

from conf import Settings


APP_NAME = 'py-app'
OTEL_AGENT_SERVICE = 'http://otel-agent-service.monitoring:4317'
APP_MONITORING_FLAG = 'APP_MONITORING'
monitoring_on = str(os.getenv(APP_MONITORING_FLAG, '1')).strip().lower() in ('on', 'yes', 'true', '1')
sys.stdout.write(f'-------{APP_MONITORING_FLAG}={monitoring_on} -----------\n')



def _set_ns():
    return Settings.NS


def get_this_runtime():
    return ".".join(
        map(
            str,
            sys.version_info[:3]
            if sys.version_info.releaselevel == "final" and not sys.version_info.serial
            else sys.version_info,
        )
    )


def get_this_namespace():
    _ns = os.getenv('MY_POD_NAMESPACE', None)
    if not _ns:
        _ns = _set_ns()
    return _ns



def set_endpoints():
    otel_metrics_endpoint = os.getenv(OTEL_EXPORTER_OTLP_ENDPOINT, '')
    otel_traces_endpoint = os.getenv(OTEL_EXPORTER_OTLP_TRACES_ENDPOINT, '')
    otel_default_endpoint = 'http://localhost:4317' if 'local' in Settings.ENV else OTEL_AGENT_SERVICE

    if not otel_metrics_endpoint:
        os.environ.setdefault(OTEL_EXPORTER_OTLP_ENDPOINT, otel_default_endpoint)
    if not otel_traces_endpoint:
        os.environ.setdefault(OTEL_EXPORTER_OTLP_TRACES_ENDPOINT, otel_default_endpoint)


def get_app_name():
    return os.getenv('APP_NAME', f"{APP_NAME}-{Settings.ENV}")

FEATURE_COUNTER: metrics.Counter | None = None
FEATURE_DURATION_HIST: metrics.Histogram | None = None

def record_feature_counter(attr: dict):
    if FEATURE_COUNTER is not None and attr:
        FEATURE_COUNTER.add(1, attributes=attr)

def record_feature_duration_histogram(duration_ms, attr: dict):
    if FEATURE_DURATION_HIST is not None and attr:
        FEATURE_DURATION_HIST.record(duration_ms, attributes=attr)


def resource_and_providers_init():
    set_endpoints()
    _runtime = get_this_runtime()
    _ns = get_this_namespace()
    # Service name is required for most backends
    #   and although it's not necessary for console export,
    #   it's good to set service name anyway.
    resource = Resource.create(attributes={
        ResAttr.CONTAINER_RUNTIME: f'{Lang.PYTHON.value}{_runtime}',
        ResAttr.CONTAINER_NAME: get_app_name(),

        ResAttr.SERVICE_NAME: get_app_name(),
        ResAttr.SERVICE_VERSION: os.getenv('IMAGE_VERSION', 'unknown-image'),
        ResAttr.SERVICE_NAMESPACE: _ns,

        ResAttr.DEPLOYMENT_ENVIRONMENT: os.getenv('APP_ENV', Settings.ENV),

        'node.ip': os.getenv('NODE_IP', 'unknown'),
        'pod': os.getenv('POD_NAME', 'unknown'),
        'pod.namespace': _ns,
    })
    # -------- Traces --------
    tracer_provider = TracerProvider(resource=resource)
    # ---------- Endpoint which points to otel-collector, is provided in $OTEL_EXPORTER_OTLP_TRACES_ENDPOINT env;
    span_processor = BatchSpanProcessor(OTLPSpanExporter())         # for deployments
    # span_processor = BatchSpanProcessor(ConsoleSpanExporter())    # for local: to read spans in console.

    # tracer_provider = None                                        # to disable tracing.
    tracer_provider.add_span_processor(span_processor)
    trace.set_tracer_provider(tracer_provider)

    # -------- Metrics --------
    # ---------- Endpoint which points to otel-collector, is provided in $OTEL_EXPORTER_OTLP_ENDPOINT env;
    reader = PeriodicExportingMetricReader(OTLPMetricExporter())        # for deployments
    # reader = PeriodicExportingMetricReader(ConsoleMetricExporter())   # for local: to read metrics in console.

    meter_provider = MeterProvider(resource=resource, metric_readers=[reader])
    # meter_provider = None                                         # to disable metrics.
    metrics.set_meter_provider(meter_provider)

    return resource, meter_provider, tracer_provider


def start_metrics_n_traces_func(fast_api_app_instance=None, is_worker=False):
    global FEATURE_COUNTER, FEATURE_DURATION_HIST
    _,  meter_provider, tracer_provider = resource_and_providers_init()

    FEATURE_COUNTER = meter_provider.get_meter(
        __name__).create_counter('app.feature.total_invoke_count',unit='request',
                                 description='Tracks the total no.of times a feature is invoked')
    FEATURE_DURATION_HIST = meter_provider.get_meter(
        __name__).create_histogram('app.feature.duration', unit='ms',
                                   description='measures the duration of the feature function')


    # Instrument redis
    RedisInstrumentor().instrument(tracer_provider=tracer_provider,
                                   meter_provider=meter_provider)

    # Instrument requests module
    RequestsInstrumentor().instrument(meter_provider=meter_provider,
                                      tracer_provide=tracer_provider)


    # Instrument the main server
    if fast_api_app_instance:
        FastAPIInstrumentor().instrument_app(fast_api_app_instance,
                                             meter_provider=meter_provider,
                                             tracer_provider=tracer_provider)


app = FastAPI()

start_metrics_n_traces_func(app)

@app.get("/ping/")
def ping_function():
    return {"result": "Pong"}



@app.get("/propagate/")
def ping_function():
    resp = requests.get('http://localhost:8080/api/ping/').json()
    return {"resultFromOtherService": resp}