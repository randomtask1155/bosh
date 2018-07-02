require 'bosh/director/api/controllers/base_controller'

module Bosh::Director
  module Api::Controllers
    class DeploymentConfigsController < BaseController
      get '/', scope: :list_configs do
        deployment_name = params['deployment']

        results = []

        deployment = Bosh::Director::Api::DeploymentManager.new.find_by_name(deployment_name)
        deployment.cloud_configs.each { |c|
          results << model_to_hash(deployment).merge(config: c)
        }

        # configs = Bosh::Director::Api::ConfigManager.new.find(
        #   type: params['type'],
        #   name: params['name'],
        #   limit: limit,
        # ).select { |config| @permission_authorizer.is_granted?(config, :read, token_scopes) }
        #
        # json_encode(configs.map { |config| sql_to_hash(config) })
        json_encode(results)
      end

      private

      def check(param, name)
        raise BadConfigRequest, "'#{name}' is required" if param[name].nil? || param[name].empty?
      end

      def model_to_hash(deployment)
        {
          id: deployment.id,
          deployment: deployment.name,
        }
      end

      def check_name_and_type(manifest, name, type)
        check(manifest, name)
        raise BadConfigRequest, "'#{name}' must be a #{type.to_s.downcase}" unless manifest[name].is_a?(type)
      end

      def validate_type_and_name(config)
        check_name_and_type(config, 'type', String)
        check_name_and_type(config, 'name', String)
      end
    end
  end
end
