# cloudnative-observability-operator

gRPC アプリを管理する Kubernetes Operator を提供するリポジトリです。

## 成果物
- CRD: GrpcBurner (v1alpha1)
- Controller/Reconciler（OwnerRef/Finalizer/GC/指数バックオフ）
- Status.Conditions + Events
- Helm Chart (crds 同梱, installCRDs トグル)
- /metrics, 構造化ログ, OTLP トレース送出

## 契約
- CRD Kind: GrpcBurner
- values.managedApp.image/tag でアプリ差し替え可能
- Helm Chart は OCI registry に公開
- Secret 名は contracts.md に準拠

## Quickstart
```bash
helm install grpcburner-operator oci://ghcr.io/YOUR_ORG/grpcburner-operator --version X.Y.Z
```

## 前提条件
- Go v1.24.0+
- Docker 17.03+.
- kubectl v1.11.3+.
- Kubernetes クラスタ v1.11.3+

## 開発、デプロイ手順(make/kubectl)
**CRDのインストール**
```bash
make install
```
**マネージャのデプロイ**
```bash
make deploy IMG=<some-registry>/cloudnative-observability-operator:tag
```
**サンプルCRの適用**
```bash
kubectl apply -k config/samples/
```
**削除(アンインストール)**
```bash
kubectl delete -k config/samples/
make uninstall
```

## MVP
- CRD v1alpha1
- Reconciler + GC + Finalizer
- Status.Conditions (Ready/Degraded)
- Helm 配布
- 観測導線 (/metrics, logs, OTLP)

## Plus
- updateStrategy / maxSurge 対応
- cosign 署名 + syft SBOM
- conversion webhook 準備

## 受け入れ基準チェックリスト
- [ ] helm install/uninstall が成功
- [ ] CR 作成 → Deployment/Service 生成
- [ ] ダミー image で Ready になる
- [ ] 異常 image で Degraded + Warning Event
- [ ] /metrics/OTLP が送出される

## スコープ外
- アプリ本体の実装
- 監視スタックの細部

## ライセンス
MIT License
