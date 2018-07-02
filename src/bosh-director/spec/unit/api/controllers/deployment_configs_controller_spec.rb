require 'spec_helper'
require 'rack/test'
require 'bosh/director/api/controllers/configs_controller'

module Bosh::Director
  describe Api::Controllers::DeploymentConfigsController do
    include Rack::Test::Methods

    subject(:app) { Api::Controllers::DeploymentConfigsController.new(config) }
    let(:config) do
      config = Config.load_hash(SpecHelper.spec_get_director_config)
      identity_provider = Support::TestIdentityProvider.new(config.get_uuid_provider)
      allow(config).to receive(:identity_provider).and_return(identity_provider)
      config
    end

    describe 'GET', '/' do
      context 'with authenticated admin user' do
        let(:deployment_name) { 'fake-dep-name' }
        let(:cloud_config) { Models::Config.make(:cloud) }
        let(:cloud_configs)  { [cloud_config] }
        let!(:deployment) do
          model = Models::Deployment.make(
            name: 'fake-dep-name',
            manifest: '---',
          )
          model.cloud_configs = cloud_configs

          model
        end

        before(:each) do
          authorize('admin', 'admin')
        end

        context 'when one deployment name is given' do
          it 'returns all active configs for that deployment' do
            get "/?deployment=#{deployment_name}"
            expect(last_response.status).to eq(200)
            deployments = JSON.parse(last_response.body)
            expect(deployments.count).to eq(1)
            expect(deployments.first['id']).to eq(deployment.id)
            expect(deployments.first['deployment']).to eq(deployment_name)
          end
        end

        context 'only lists active configs' do
          let(:named_cloud_config) { Models::Config.make(:cloud, name: 'custom-name') }
          let(:cloud_configs)  { [cloud_config, named_cloud_config] }

          it 'returns one result per config' do
            get "/?deployment=#{deployment_name}"
            expect(last_response.status).to eq(200)
            deployments = JSON.parse(last_response.body)
            expect(deployments.count).to eq(2)
          end
        end
      end
    end
  end
end